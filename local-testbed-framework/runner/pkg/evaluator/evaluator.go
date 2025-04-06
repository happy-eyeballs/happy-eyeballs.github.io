package evaluator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/environment"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/result"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/script"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/target"
	"golang.org/x/exp/slices"
	"path"
)

const evaluationTargetTag = "_evaluation"

type Evaluator interface {
	Evaluate(ctx context.Context, logger zerolog.Logger, env environment.Environment, artifactsDirectoryPath string, result result.Result)
}

type evaluator struct {
	scriptPath           string
	target               *target.Target
	workingDirectoryPath string
}

func NewEvaluator(targets []*target.Target, evaluationScriptPath string, testCaseName string) (Evaluator, error) {
	var targetToRunEvaluationOn *target.Target
	for _, currentTarget := range targets {
		if slices.Contains(currentTarget.Tags(), evaluationTargetTag) {
			targetToRunEvaluationOn = currentTarget
			break
		}
	}

	if targetToRunEvaluationOn == nil {
		return nil, fmt.Errorf("special target with tag %q for evaluation scripts does not exist", evaluationTargetTag)
	}

	return &evaluator{
		scriptPath:           evaluationScriptPath,
		target:               targetToRunEvaluationOn,
		workingDirectoryPath: path.Join(targetToRunEvaluationOn.Dir(testCaseName), evaluationTargetTag),
	}, nil
}

func (e *evaluator) Evaluate(ctx context.Context, logger zerolog.Logger, env environment.Environment, artifactsDirectoryPath string, result result.Result) {
	logger.Debug().Msg("Test case evaluation started")
	defer logger.Debug().Msg("Test case evaluation finished")

	evaluationScript := script.NewScript(script.Config{
		Script:    e.scriptPath,
		Artifacts: []string{},
		Always:    false,
	})

	artifactsCopyPath := path.Join(e.workingDirectoryPath, "_artifacts")

	err := e.target.CopyToTarget(artifactsDirectoryPath, artifactsCopyPath)
	if err != nil {
		logger.Err(err).Msg("Could not copy artifacts to evaluation target")
		return
	}

	stdout := evaluationScript.ExecuteScript(ctx, logger, e.target,
		env.With("EVALUATION_ARTIFACTS_DIR", artifactsCopyPath), e.workingDirectoryPath)

	var outputs []struct {
		IterationVariable int     `json:"iteration_variable"`
		Measurement       string  `json:"measurement"`
		Value             *string `json:"value"`
	}
	err = json.Unmarshal(stdout, &outputs)
	if err != nil {
		logger.Err(err).Msg("Error parsing evaluation script output to JSON")
		return
	}

	for _, output := range outputs {
		err := result.
			WithIterationVariable(output.IterationVariable).
			WithMeasurement(output.Measurement).
			Record(output.Value)
		if err != nil {
			logger.Err(err).Msg("Error recording result")
			continue
		}
	}
}
