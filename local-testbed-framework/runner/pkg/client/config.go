package client

import (
	"errors"
	"fmt"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/utils"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Versions *VersionsConfig `yaml:"versions,omitempty"`
}

type VersionsConfig struct {
	EnvironmentVariableName string   `yaml:"environmentVariableName"`
	Values                  []string `yaml:"values"`
}

func ParseClientConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	data, err := utils.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing contents of config: %w", err)
	}

	return &config, nil
}
