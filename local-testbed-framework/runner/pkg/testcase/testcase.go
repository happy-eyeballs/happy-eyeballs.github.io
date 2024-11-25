package testcase

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/environment"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/evaluator"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/result"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/script"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/stage"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/target"
	"path"
	"strconv"
	"sync"
)

const MetaTestCaseName = "_all"

type TestCase interface {
	Run(ctx context.Context, logger zerolog.Logger, targets []*target.Target, env environment.Environment, artifactsDirectoryPath string)
	RunStage(ctx context.Context, logger zerolog.Logger, targets []*target.Target, env environment.Environment, artifactsDirectoryPath string, stageName string)
	Evaluate(ctx context.Context, logger zerolog.Logger, targets []*target.Target, env environment.Environment, artifactsDirectoryPath string, result result.Result)
	Name() string
}

type testCase struct {
	name             string
	stages           map[string]stage.Stage
	stageOrder       []string
	repetitionConfig *RepetitionConfig
	evaluationConfig *EvaluationConfig
}

func NewTestCase(name string, config *Config) (TestCase, error) {
	stages := make(map[string]stage.Stage)

	for _, stageName := range config.Stages {
		scriptsPerTargetTag := make(map[string][]script.Script)

		for _, targetConfig := range config.Targets {
			newScripts, err := targetConfig.GetScriptsForStage(stageName)
			if err != nil {
				return nil, err
			}

			if len(newScripts) == 0 {
				continue
			}

			currentScripts, ok := scriptsPerTargetTag[targetConfig.Tag]
			if !ok {
				scriptsPerTargetTag[targetConfig.Tag] = newScripts
			} else {
				currentScripts = append(currentScripts, newScripts...)
			}
		}

		stages[stageName] = stage.NewStage(scriptsPerTargetTag)
	}

	if config.Repeat != nil {
		if err := config.Repeat.validate(); err != nil {
			return nil, fmt.Errorf("invalid test case repetition config: %w", err)
		}
	}

	return &testCase{
		name:             name,
		stages:           stages,
		stageOrder:       config.Stages,
		repetitionConfig: config.Repeat,
		evaluationConfig: config.Evaluation,
	}, nil
}

func (t *testCase) Run(ctx context.Context, logger zerolog.Logger, targets []*target.Target, env environment.Environment, artifactsDirectoryPath string) {
	if t.repetitionConfig == nil {
		t.runAllStages(ctx, logger, targets, env, artifactsDirectoryPath)
		return
	}

	for value := t.repetitionConfig.From; value <= t.repetitionConfig.To; value += t.repetitionConfig.Step {
		formattedValue := strconv.FormatInt(int64(value), 10)

		env := env.With(t.repetitionConfig.EnvironmentVariableName, formattedValue)
		logger := logger.With().Str(t.repetitionConfig.EnvironmentVariableName, formattedValue).Logger()
		artifactsDirectoryPath := path.Join(artifactsDirectoryPath, formattedValue)

		t.runAllStages(ctx, logger, targets, env, artifactsDirectoryPath)
	}
}

func (t *testCase) runAllStages(ctx context.Context, logger zerolog.Logger, targets []*target.Target, env environment.Environment, artifactsDirectoryPath string) {
	for _, stageName := range t.stageOrder {
		t.RunStage(ctx, logger, targets, env, artifactsDirectoryPath, stageName)
	}
}

func (t *testCase) RunStage(ctx context.Context, logger zerolog.Logger, targets []*target.Target, env environment.Environment, artifactsDirectoryPath string, stageName string) {
	logger = logger.With().Str("stage", stageName).Logger()
	logger.Debug().Msg("Stage execution started")

	stageToRun, ok := t.stages[stageName]
	if !ok {
		logger.Error().Msg("Unknown stage")
		return
	}

	var wg sync.WaitGroup

	for _, currentTarget := range targets {
		logger := logger.With().Str("target", currentTarget.DisplayName()).Logger()
		logger.Debug().Msg("Started execution on target")

		wg.Add(1)

		go func(target *target.Target) {
			stageToRun.RunOnTarget(ctx, logger, target, env, t.getWorkingDirectory(target), artifactsDirectoryPath)
			wg.Done()
		}(currentTarget)
	}

	wg.Wait()

	logger.Debug().Msg("Stage execution finished")
}

func (t *testCase) Evaluate(ctx context.Context, logger zerolog.Logger, targets []*target.Target, env environment.Environment, artifactsDirectoryPath string, result result.Result) {
	if t.evaluationConfig == nil {
		return
	}

	testCaseEvaluator, err := evaluator.NewEvaluator(targets, t.evaluationConfig.Script, t.name)
	if err != nil {
		logger.Err(err).Msg("Error instantiating test case evaluator")
		return
	}

	testCaseEvaluator.Evaluate(ctx, logger, env, artifactsDirectoryPath, result.WithTestCaseName(t.name))
}

func (t *testCase) getWorkingDirectory(target *target.Target) string {
	return target.Dir(t.name)
}

func (t *testCase) Name() string {
	return t.name
}
