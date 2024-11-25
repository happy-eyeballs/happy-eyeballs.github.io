package client

import (
	"fmt"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/environment"
)

type Client interface {
	DirectoryName() string
	DisplayName() string
	EnvironmentWithVersion(env environment.Environment) environment.Environment
}

type client struct {
	directoryName string

	version                    string
	versionEnvironmentVariable string
}

func NewClient(directoryName string, version string, versionEnvironmentVariable string) Client {
	return &client{
		directoryName:              directoryName,
		version:                    version,
		versionEnvironmentVariable: versionEnvironmentVariable,
	}
}

func (c *client) DirectoryName() string {
	return c.directoryName
}

func (c *client) DisplayName() string {
	if c.version != "" {
		return fmt.Sprintf("%s-%s", c.directoryName, c.version)
	}

	return c.directoryName
}

func (c *client) EnvironmentWithVersion(env environment.Environment) environment.Environment {
	if c.versionEnvironmentVariable == "" || c.version == "" {
		return env
	}

	return env.With(c.versionEnvironmentVariable, c.version)
}
