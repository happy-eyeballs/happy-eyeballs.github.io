package testcase

import (
	"fmt"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/script"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/utils"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Stages     []string          `yaml:"stages"`
	Targets    []TargetConfig    `yaml:"targets"`
	Repeat     *RepetitionConfig `yaml:"repeat,omitempty"`
	Evaluation *EvaluationConfig `yaml:"evaluation,omitempty"`
}

type TargetConfig struct {
	Tag     string          `yaml:"tag"`
	Scripts []script.Config `yaml:"scripts"`
}

type RepetitionConfig struct {
	EnvironmentVariableName string `yaml:"environmentVariableName"`
	From                    int    `yaml:"from"`
	To                      int    `yaml:"to"`
	Step                    int    `yaml:"step"`
}

type EvaluationConfig struct {
	Script string `yaml:"script"`
}

func ParseTestCaseConfig(path string) (*Config, error) {
	data, err := utils.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing contents of config: %w", err)
	}

	if len(config.Stages) <= 0 || len(config.Targets) <= 0 {
		return nil, fmt.Errorf("stages or targets are missing in config at %q", path)
	}

	return &config, nil
}

func (c TargetConfig) GetScriptsForStage(stageName string) ([]script.Script, error) {
	scripts := make([]script.Script, 0)

	for _, scriptForTarget := range c.Scripts {
		if scriptForTarget.Stage == stageName {
			scripts = append(scripts, script.NewScript(scriptForTarget))
		}
	}

	return scripts, nil
}

func (c *RepetitionConfig) validate() error {
	if c.EnvironmentVariableName == "" {
		return fmt.Errorf("environment variable name must be set")
	}

	if c.Step == 0 {
		return fmt.Errorf("step size may not be 0")
	}

	if c.Step > 0 && c.To < c.From {
		return fmt.Errorf("for a positive step size, 'to' has to be greater than 'from'")
	}

	if c.Step < 0 && c.To > c.From {
		return fmt.Errorf("for a negative step size, 'to' has to be less than 'from'")
	}

	return nil
}
