package script

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/environment"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/executor/execcontext"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/target"
	"path"
	"strings"
	"time"
)

type Script interface {
	Run(ctx context.Context, logger zerolog.Logger, target *target.Target, env environment.Environment, workingDirectoryPath string, artifactsDirectoryPath string)
	ExecuteScript(ctx context.Context, logger zerolog.Logger, target *target.Target, env environment.Environment, workingDirectoryPath string) (stdout []byte)
}

type script struct {
	script    string
	artifacts []string
	always    bool
}

func NewScript(config Config) Script {
	return &script{
		script:    config.Script,
		artifacts: config.Artifacts,
		always:    config.Always,
	}
}

func (s *script) Run(ctx context.Context, logger zerolog.Logger, target *target.Target, env environment.Environment, workingDirectoryPath string, artifactsDirectoryPath string) {
	hasErrorOccurred := errors.Is(ctx.Err(), context.Canceled)
	if hasErrorOccurred && !s.always {
		logger.Warn().Msg("Skipping script due to earlier failures")
		return
	}

	s.ExecuteScript(ctx, logger, target, env, workingDirectoryPath)
	s.collectArtifactsFromTarget(logger, target, workingDirectoryPath, artifactsDirectoryPath)
}

func (s *script) ExecuteScript(ctx context.Context, logger zerolog.Logger, target *target.Target, env environment.Environment, workingDirectoryPath string) (stdout []byte) {
	if s.script == "" {
		return
	}

	commandCtx := ctx
	if s.always {
		commandCtx = context.Background() // otherwise the script execution is aborted immediately
	}

	scriptPath := path.Join(workingDirectoryPath, s.script)
	result, err := target.Exec(commandCtx, env.With("BASE_DIR", workingDirectoryPath), scriptPath)

	outputLoggerBuilder := logger.With().
		Str("target", target.DisplayName()).
		Str("script", scriptPath)

	if result != nil {
		stdout = result.Stdout
		executionTime := time.Duration(result.EndTime-result.StartTime) * time.Nanosecond

		outputLoggerBuilder = outputLoggerBuilder.
			Str("executionTime", executionTime.String()).
			Str("stdout", strings.TrimSpace(string(result.Stdout))).
			Str("stderr", strings.TrimSpace(string(result.Stderr))).
			Int("exitStatus", result.ExitStatus)
	}

	outputLogger := outputLoggerBuilder.Logger()

	if err != nil {
		outputLogger.Err(err).Msg("Script execution failed")
		execcontext.CancelExecContext(ctx)
		return
	}

	outputLogger.Info().Msg("Script execution successful")
	return
}

func (s *script) collectArtifactsFromTarget(logger zerolog.Logger, target *target.Target, workingDirectoryPath string, artifactsDirectoryPath string) {
	if s.artifacts == nil || len(s.artifacts) == 0 {
		return
	}

	for _, artifact := range s.artifacts {
		srcPath := path.Join(workingDirectoryPath, artifact)
		dstPath := path.Join(artifactsDirectoryPath, artifact)

		err := target.CopyFromTarget(srcPath, dstPath)
		if err != nil {
			logger.Err(err).Str("artifact", artifact).Msg("Could not collect artifact(s) from target")
			return
		}
	}
}
