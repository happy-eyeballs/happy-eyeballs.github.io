package utils

import (
	"errors"
	"fmt"
	"golang.org/x/term"
	"io/fs"
	"os"
	"path"
)

func GetAbsPath(ps ...string) (string, error) {
	p := path.Join(ps...) // Cleans path
	if path.IsAbs(p) {
		return p, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current working directory: %w", err)
	}

	return path.Join(cwd, p), nil
}

func ReadSecretFromStdin(msg string) ([]byte, error) {
	fmt.Print(msg)
	secret, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // Print new line (ENTER not printed to terminal)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret from terminal: %s", err)
	}
	return secret, nil
}

func CreateFile(filePath string) (*os.File, error) {
	if err := os.MkdirAll(path.Dir(filePath), fs.ModeDir|fs.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create directory: %v", err)
	}
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create file: %v", err)
	}
	return file, nil
}

func ReadFile(path string) ([]byte, error) {
	configFile, err := GetAbsPath(path)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("file \"%s\" could not be found", path)
	}
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not read file \"%s\": %v", path, err)
	}
	return data, nil
}
