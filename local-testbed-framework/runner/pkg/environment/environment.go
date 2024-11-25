package environment

type Environment interface {
	With(name string, value string) Environment
	GetVariables() map[string]string
}

type environment struct {
	variables map[string]string
}

func NewEnvironment() Environment {
	return &environment{
		variables: make(map[string]string),
	}
}

func (e *environment) With(key string, value string) Environment {
	variablesCopy := make(map[string]string, len(e.variables))
	for k, v := range e.variables {
		variablesCopy[k] = v
	}

	variablesCopy[key] = value
	return &environment{variables: variablesCopy}
}

func (e *environment) GetVariables() map[string]string {
	return e.variables
}
