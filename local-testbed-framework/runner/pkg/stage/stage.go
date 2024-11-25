package stage

import (
	"context"
	"github.com/rs/zerolog"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/environment"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/script"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/target"
	"path"
	"sync"
)

const (
	BeforeAll  = "beforeAll"
	AfterAll   = "afterAll"
	BeforeEach = "beforeEach"
	AfterEach  = "afterEach"
)

type Stage interface {
	RunOnTarget(ctx context.Context, logger zerolog.Logger, target *target.Target, env environment.Environment, workingDirectoryPath string, artifactsDirectoryPath string)
}

type stage struct {
	scriptsPerTargetTag map[string][]script.Script
}

func NewStage(scriptsPerTargetTag map[string][]script.Script) Stage {
	return &stage{
		scriptsPerTargetTag: scriptsPerTargetTag,
	}
}

func (s *stage) RunOnTarget(ctx context.Context, logger zerolog.Logger, target *target.Target, env environment.Environment, workingDirectoryPath string, artifactsDirectoryPath string) {
	var wg sync.WaitGroup

	for _, tag := range target.Tags() {
		scripts, ok := s.scriptsPerTargetTag[tag]
		if !ok {
			continue
		}

		for _, currentScript := range scripts {
			wg.Add(1)

			go func(tag string, script script.Script) {
				script.Run(ctx, logger, target, env, path.Join(workingDirectoryPath, tag), path.Join(artifactsDirectoryPath, tag))
				wg.Done()
			}(tag, currentScript)
		}
	}

	wg.Wait()
}
