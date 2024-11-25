package cmd

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/cmd/common"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/cmd/internal"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/client"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/environment"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/executor"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/report"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/result"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/resultdatabase"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/target"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/testcase"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/utils"
	"os"
	"os/signal"
	"path"
	"strings"
	"time"
)

type ExecOptions struct {
	ConfigFile             string
	InteractiveClient      bool
	InteractiveServer      bool
	ArtifactsDirectoryPath string
	TestCasesDirectoryPath string
	SelectedTestCases      string
	ClientsDirectoryPath   string
	SelectedClients        string
}

var execOptions ExecOptions
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute arbitrary test cases",
	Long: `
Automatically execute a user-defined list of test cases. 
Tests are simultaneously executed on client and server. 
Stages within test cases are used as synchronization barriers.

A non-zero exit status signals a fatal error during execution of the test
cases from which the runner was not able to recover.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdExec()
	},
}

func init() {
	execCmd.Flags().StringVarP(
		&execOptions.ConfigFile,
		"config", "c", "config.yml",
		"Path to the configuration file",
	)
	execCmd.Flags().StringVarP(
		&execOptions.ArtifactsDirectoryPath,
		"artifacts-directory-path", "a", "artifacts",
		"Path to the directory that artifacts should be written to",
	)
	execCmd.Flags().StringVarP(
		&execOptions.TestCasesDirectoryPath,
		"test-cases-directory-path", "d", "testcases",
		"Path to the directory storing all available test cases",
	)
	execCmd.Flags().StringVarP(
		&execOptions.SelectedTestCases,
		"tests", "t", "*",
		"Comma-seperated list of test cases",
	)
	execCmd.Flags().StringVar(
		&execOptions.ClientsDirectoryPath,
		"clients-directory-path", "clients",
		"Path to the directory storing all available clients",
	)
	execCmd.Flags().StringVar(
		&execOptions.SelectedClients,
		"clients", "*",
		"Comma-seperated list of clients",
	)

	rootCmd.AddCommand(execCmd)
}

func cmdExec() {
	logger := internal.NewLogger()

	err := run(logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error(s) during execution of program. Exiting with status code 1")
		os.Exit(1)
		return
	}

	logger.Info().Msg("Exiting program")
	os.Exit(0)
}

func run(logger zerolog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	config, err := common.ParseConfig(execOptions.ConfigFile)
	if err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	testCases, err := loadTestCases(execOptions.TestCasesDirectoryPath)
	if err != nil {
		return fmt.Errorf("error loading test cases: %w", err)
	}

	clients, err := loadClients(execOptions.ClientsDirectoryPath)
	if err != nil {
		return fmt.Errorf("error loading clients: %w", err)
	}

	targets, err := loadTargets(logger, config)
	if err != nil {
		return fmt.Errorf("error loading targets: %w", err)
	}

	selectedTestCases := executor.FilterTestCases(testCases, strings.Split(execOptions.SelectedTestCases, ","))
	selectedClients := executor.FilterClients(clients, strings.Split(execOptions.SelectedClients, ","))

	logger.Info().Msgf("Selected test cases: %d", len(selectedTestCases)-1)
	logger.Info().Msgf("Selected clients: %d", len(selectedClients))

	worker, err := executor.NewExecutor(logger, targets, selectedTestCases, selectedClients,
		execOptions.TestCasesDirectoryPath, execOptions.ClientsDirectoryPath)
	if err != nil {
		return fmt.Errorf("error instantiating executor: %w", err)
	}

	resultDatabase, err := resultdatabase.NewResultDatabase(path.Join(execOptions.ArtifactsDirectoryPath, "results.db"))
	if err != nil {
		return fmt.Errorf("error creating result database: %w", err)
	}

	timestamp := time.Now()
	runId := int(timestamp.Unix())
	artifactsDirectoryPath := path.Join(execOptions.ArtifactsDirectoryPath, timestamp.Format("20060102_150405"))

	logger.Info().Int("runId", runId).Msg("Starting run")
	worker.Run(
		ctx,
		artifactsDirectoryPath,
		environment.NewEnvironment(),
		result.NewResult(resultDatabase).WithRunId(runId),
	)
	logger.Info().Msg("Ending run")

	report.NewReport().GenerateReport(logger, artifactsDirectoryPath, resultDatabase, runId)

	return nil
}

func loadTestCases(testCasesDirectoryPath string) ([]testcase.TestCase, error) {
	absoluteTestCasesDirectoryPath, err := utils.GetAbsPath(testCasesDirectoryPath)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(absoluteTestCasesDirectoryPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read test cases directory: %w", err)
	}

	testCases := make([]testcase.TestCase, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		testCaseConfigPath := path.Join(testCasesDirectoryPath, entry.Name(), "config.yml")
		testCaseConfig, err := testcase.ParseTestCaseConfig(testCaseConfigPath)
		if err != nil {
			return nil, fmt.Errorf("error parsing test case config at %q: %w", testCaseConfigPath, err)
		}

		testCase, err := testcase.NewTestCase(entry.Name(), testCaseConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating test case: %w", err)
		}

		testCases = append(testCases, testCase)
	}

	return testCases, nil
}

func loadClients(clientsDirectoryPath string) ([]client.Client, error) {
	absoluteClientsDirectoryPath, err := utils.GetAbsPath(clientsDirectoryPath)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(absoluteClientsDirectoryPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read clients directory: %w", err)
	}

	clients := make([]client.Client, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		clientConfigPath := path.Join(clientsDirectoryPath, entry.Name(), "config.yml")
		clientConfig, err := client.ParseClientConfig(clientConfigPath)
		if err != nil {
			return nil, fmt.Errorf("error parsing client config at %q: %w", clientConfigPath, err)
		}

		if clientConfig == nil || clientConfig.Versions == nil || len(clientConfig.Versions.Values) == 0 {
			clients = append(clients, client.NewClient(entry.Name(), "", ""))
			continue
		}

		for _, clientVersion := range clientConfig.Versions.Values {
			clients = append(clients, client.NewClient(entry.Name(), clientVersion, clientConfig.Versions.EnvironmentVariableName))
		}
	}

	return clients, nil
}

func loadTargets(logger zerolog.Logger, config *common.Config) ([]*target.Target, error) {
	targets := make([]*target.Target, 0)

	for _, targetConf := range config.Targets {
		logger.Info().Str("host", targetConf.Name).Msg("Initializing target")

		newTarget, err := target.NewTarget(&targetConf, logger)
		if err != nil {
			logger.Err(err).Str("host", targetConf.Name).Msg("Failed to create target")
			return nil, err
		}

		targets = append(targets, newTarget)
	}

	return targets, nil
}
