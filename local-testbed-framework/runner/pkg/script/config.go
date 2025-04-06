package script

type Config struct {
	Stage     string   `yaml:"stage"`
	Always    bool     `yaml:"always,omitempty"`
	Script    string   `yaml:"script,omitempty"`
	Artifacts []string `yaml:"artifacts,omitempty"`
}
