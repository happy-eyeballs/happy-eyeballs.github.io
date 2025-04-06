package common

import (
	"fmt"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/utils"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Targets []HostConfig
}

type HostConfig struct {
	Name              string
	SSH               *SSHConfig // optional
	EnvironmentPrefix string     `yaml:"environmentPrefix"`
	Environment       []string   // optional
	Tags              []string
}

type SSHConfig struct {
	Hostname   string
	User       string
	Port       int16
	Privkey    string
	Knownhosts string
}

func ParseConfig(path string) (*Config, error) {
	data, err := utils.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := Config{}
	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing contents of configuration file: %v", err)
	}

	if config.Targets == nil {
		return nil, fmt.Errorf("target section missing in configuration file %q", path)
	}

	return &config, nil
}
