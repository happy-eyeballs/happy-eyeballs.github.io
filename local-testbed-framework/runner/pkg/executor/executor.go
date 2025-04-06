package executor

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/client"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/environment"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/executor/execcontext"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/result"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/stage"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/target"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/testcase"
	"golang.org/x/exp/slices"
	"path"
)

type Executor interface {
	Run(ctx context.Context, artifactsDirectoryPath string, env environment.Environment, result result.Result)
}

type executor struct {
	logger zerolog.Logger

	targets []*target.Target
	clients []client.Client

	metaTestCase testcase.TestCase
	testCases    []testcase.TestCase

	testCasesDirectoryPath string
	clientsDirectoryPath   string
}

func NewExecutor(
	logger zerolog.Logger,
	targets []*target.Target,
	testCases []testcase.TestCase,
	clients []client.Client,
	testCasesDirectoryPath string,
	clientsDirectoryPath string,
) (Executor, error) {
	var metaTestCase testcase.TestCase

	for i, testCase := range testCases {
		if testCase.Name() == testcase.MetaTestCaseName {
			metaTestCase = testCase
			testCases = slices.Delete(testCases, i, i+1)
			break
		}
	}

	if metaTestCase == nil {
		return nil, fmt.Errorf("missing meta test case (%q)", testcase.MetaTestCaseName)
	}

	return &executor{
		logger:                 logger,
		targets:                targets,
		clients:                clients,
		metaTestCase:           metaTestCase,
		testCases:              testCases,
		testCasesDirectoryPath: testCasesDirectoryPath,
		clientsDirectoryPath:   clientsDirectoryPath,
	}, nil
}

func (e *executor) Run(ctx context.Context, artifactsDirectoryPath string, env environment.Environment, result result.Result) {
	for _, currentClient := range e.clients {
		logger := e.logger.With().Str("client", currentClient.DisplayName()).Logger()

		err := e.initTargetsForClient(currentClient)
		if err != nil {
			logger.Err(err).Msg("Error initializing targets")
			continue
		}

		artifactsDirectoryPath := path.Join(artifactsDirectoryPath, currentClient.DisplayName())

		logger.Info().Msg("Starting execution on client")
		e.runTestCases(ctx, logger, artifactsDirectoryPath, currentClient.EnvironmentWithVersion(env), result.WithClientName(currentClient.DisplayName()))
		logger.Info().Msg("Finished execution on client")
	}

	e.cleanupTargets()
}

func (e *executor) runTestCases(ctx context.Context, logger zerolog.Logger, artifactsDirectoryPath string, env environment.Environment, result result.Result) {
	executionCtx := execcontext.NewExecContext(ctx)

	e.metaTestCase.RunStage(executionCtx, logger, e.targets, env, artifactsDirectoryPath, stage.BeforeAll)

	for _, testCase := range e.testCases {
		env := env.With("TEST_CASE", testCase.Name())
		artifactsDirectoryPath := path.Join(artifactsDirectoryPath, testCase.Name())

		e.metaTestCase.RunStage(executionCtx, logger, e.targets, env, artifactsDirectoryPath, stage.BeforeEach)
		testCase.Run(executionCtx, logger, e.targets, env, artifactsDirectoryPath)
		e.metaTestCase.RunStage(executionCtx, logger, e.targets, env, artifactsDirectoryPath, stage.AfterEach)

		testCase.Evaluate(executionCtx, logger, e.targets, env, artifactsDirectoryPath, result)
	}

	e.metaTestCase.RunStage(executionCtx, logger, e.targets, env, artifactsDirectoryPath, stage.AfterAll)
}

func (e *executor) initTargetsForClient(client client.Client) error {
	allTestCases := append(e.testCases, e.metaTestCase)

	for _, currentTarget := range e.targets {
		for _, testCase := range allTestCases {
			err := currentTarget.InitTestCase(testCase.Name(), e.testCasesDirectoryPath)
			if err != nil {
				return fmt.Errorf("error initializing test case on target: %w", err)
			}
		}

		err := currentTarget.InitClient(client.DirectoryName(), e.clientsDirectoryPath)
		if err != nil {
			return fmt.Errorf("error initializing client on target: %w", err)
		}
	}

	return nil
}

func (e *executor) cleanupTargets() {
	for _, currentTarget := range e.targets {
		logger := e.logger.With().Str("target", currentTarget.DisplayName()).Logger()

		err := currentTarget.Done()
		if err != nil {
			logger.Err(err).Msg("Error deleting temporary directories on target")
		}

		err = currentTarget.Close()
		if err != nil {
			logger.Err(err).Msg("Error closing connection to target")
		}
	}
}
