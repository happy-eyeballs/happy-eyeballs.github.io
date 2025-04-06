package target

import (
	"bytes"
	"context"
	"fmt"
	krfs "github.com/kr/fs"
	"github.com/pkg/sftp"
	"github.com/rs/zerolog"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/cmd/common"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/environment"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/utils"
	"golang.org/x/crypto/ssh"
	"golang.org/x/exp/slices"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const ClientTargetTag = "_client"

type Target struct {
	client      *ssh.Client
	sftp        *sftp.Client
	targetDir   string
	environment map[string]string
	logger      zerolog.Logger
	config      common.HostConfig
}

type ExecutionResult struct {
	Stdout     []byte
	Stderr     []byte
	StartTime  int64
	EndTime    int64
	ExitStatus int
}

func NewTarget(config *common.HostConfig, logger zerolog.Logger) (*Target, error) {
	t := &Target{
		config:      *config,
		logger:      logger.With().Str("host", config.Name).Logger(),
		environment: make(map[string]string),
	}

	// Establish SSH connection if needed
	if config.SSH != nil {
		err := t.connectSSH()
		if err != nil {
			return nil, err
		}
		t.logger.Debug().Msg("Successfully established an SSH and SFTP connection to target")
	}

	// Create temporary directory
	result, err := t.Exec(context.Background(), nil, "mktemp", "-d")
	if err != nil {
		t.Close()
		return nil, fmt.Errorf("unable to create temporary directory: %s (%s)", result.Stderr, err)
	}

	t.targetDir = strings.TrimSpace(string(result.Stdout[:]))
	t.logger.Debug().Str("tmp", t.targetDir).Msg("Created temporary directory")

	// Populate environment variables
	if len(t.config.Environment) > 0 {
		for _, env := range t.config.Environment {
			if env := strings.SplitN(env, "=", 2); len(env) >= 2 {
				t.environment[t.envName(env[0])] = env[1]
			}
		}
	}

	return t, nil
}

func (t *Target) envName(name string) string {
	return fmt.Sprintf("%s%s", t.config.EnvironmentPrefix, name)
}

func (t *Target) Dir(suffix string) string {
	return path.Join(t.targetDir, suffix)
}

func (t *Target) DisplayName() string {
	return t.config.Name
}

func (t *Target) Tags() []string {
	return t.config.Tags
}

func (t *Target) InitTestCase(testCaseName string, testCasesDirectoryPath string) error {
	for _, tag := range t.Tags() {
		srcDir := path.Join(testCasesDirectoryPath, testCaseName, tag)
		//if tag == ClientTargetTag {
		//	srcDir = path.Join(clientsDirectoryPath, clientName)
		//}

		dstDir := path.Join(t.targetDir, testCaseName, tag)

		// Skip tag if there are no definitions for it in the test case
		_, err := os.Stat(srcDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return fmt.Errorf("error while collecting file statistics: %w", err)
		}

		err = t.CopyToTarget(srcDir, dstDir)
		if err != nil {
			return fmt.Errorf("error while copying test cases: %w", err)
		}
	}

	return nil
}

func (t *Target) InitClient(clientDirectoryName string, clientsDirectoryPath string) error {
	if !slices.Contains(t.Tags(), ClientTargetTag) {
		return nil
	}

	srcDir := path.Join(clientsDirectoryPath, clientDirectoryName)
	dstDir := path.Join(t.targetDir, ClientTargetTag)

	err := t.CopyToTarget(srcDir, dstDir)
	if err != nil {
		return fmt.Errorf("error while copying client: %w", err)
	}

	t.environment[t.envName("CLIENT_DIR")] = dstDir

	return nil
}

func (t *Target) Done() error {
	// Delete test cases from target
	if strings.HasPrefix(t.targetDir, "/tmp/") || strings.HasPrefix(t.targetDir, "/var/folders/") {
		_, err := t.Exec(context.Background(), nil, "rm", "-rf", t.targetDir)
		return err
	}
	return nil
}

func (t *Target) connectSSH() error {
	// Do nothing if no config exists
	if t.config.SSH == nil {
		return nil
	}

	client, err := connectSSH(t.config.SSH)
	if err != nil {
		return fmt.Errorf("unable to establish ssh connection: %s", err)
	}
	t.client = client

	sftp, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("unable to create sftp connection: %s", err)
	}
	t.sftp = sftp

	return nil
}

func (t *Target) openSSHSession() (*ssh.Session, error) {
	session, err := t.client.NewSession()
	if err == nil {
		return session, nil
	}

	t.logger.Error().Err(err).Msg("Could not create a new SSH session")
	t.logger.Info().Msg("SSH socket may be broken, attempting reconnect")

	err = t.Close()
	if err != nil {
		t.logger.Error().Err(err).Msg("Error occurred when closing SSH connection")
	}

	time.Sleep(time.Second)

	err = t.connectSSH()
	if err != nil {
		return nil, err
	}

	// Retry connection attempt
	session, err = t.client.NewSession()
	if err != nil {
		return nil, err
	}
	t.logger.Info().Msg("Reconnect to SSH server was successful")
	return session, nil
}

func (t *Target) Close() error {
	if t.sftp != nil {
		err := t.sftp.Close()
		if err != nil {
			return err
		}
	}
	if t.client != nil {
		err := t.client.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Target) CopyToTarget(srcPath, dstPath string) error {
	absSrcPath, err := utils.GetAbsPath(srcPath)
	if err != nil {
		return fmt.Errorf("error getting absolute src path: %w", err)
	}

	err = filepath.WalkDir(absSrcPath, func(p string, d fs.DirEntry, err error) error {
		if err != nil { // Prevent panic
			return err
		}
		srcFileInfo, err := d.Info()
		if err != nil {
			return err
		}
		if srcFileInfo.Mode()&(os.ModeSymlink|os.ModeDir) == 0 { // Skip dirs, symlinks
			dstFile := path.Join(dstPath, p[len(absSrcPath):])
			if err = t.copyFileToTarget(p, dstFile); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to copy all files to target directory: %s", err)
	}
	return nil
}

func (t *Target) CopyFromTarget(srcPath, dstPath string) error {
	var walker *krfs.Walker
	var absSrcPath string
	var err error
	if t.sftp == nil { // Local
		absSrcPath, err = utils.GetAbsPath(srcPath)
		walker = krfs.Walk(absSrcPath)
	} else { // SSH
		absSrcPath = srcPath
		walker = t.sftp.Walk(absSrcPath)
	}

	if err != nil {
		return err
	}

	for walker.Step() {
		if err := walker.Err(); err != nil {
			return fmt.Errorf("unable to copy all files to target directory: %s", err)
		}
		if walker.Stat().Mode()&(os.ModeSymlink|os.ModeDir) == 0 { // Skip dirs, symlinks
			srcFile := walker.Path()

			dstFile, err := utils.GetAbsPath(dstPath, srcFile[len(absSrcPath):])
			if err != nil {
				return fmt.Errorf("error getting absolute dst path: %w", err)
			}

			if err := t.copyFileFromTarget(srcFile, dstFile); err != nil {
				return fmt.Errorf("unable to copy all files to target directory: %w", err)
			}
		}
	}
	return nil
}

func (t *Target) copyFileToTarget(srcPath, dstPath string) error {
	var srcFile, dstFile File
	var err error

	t.logger.Debug().
		Str("srcPath", srcPath).
		Str("dstPath", dstPath).
		Msg("Copying")

	absSrcPath, err := utils.GetAbsPath(srcPath)
	if err != nil {
		return fmt.Errorf("error getting absolute src path: %w", err)
	}

	srcFile, err = os.Open(absSrcPath)
	if err != nil {
		return fmt.Errorf("unable to open source file: %v", err)
	}

	defer srcFile.Close()

	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("unable to read metadata of source file: %v", err)
	}

	if t.sftp == nil { // Local
		dstPath, err := utils.GetAbsPath(dstPath)
		if err != nil {
			return fmt.Errorf("error getting absolute dst path: %w", err)
		}

		if err = os.MkdirAll(path.Dir(dstPath), fs.ModeDir|fs.ModePerm); err != nil {
			return fmt.Errorf("unable to create destination directory: %v", err)
		}
		dstFile, err = os.Create(dstPath)
	} else { // SSH
		if err = t.sftp.MkdirAll(path.Dir(dstPath)); err != nil {
			return fmt.Errorf("unable to create destination directory: %v", err)
		}
		dstFile, err = t.sftp.Create(dstPath)
	}
	if err != nil {
		return fmt.Errorf("unable to create destination file: %v", err)
	}
	defer dstFile.Close()
	dstFile.Chmod(srcFileInfo.Mode())

	if _, err = dstFile.ReadFrom(srcFile); err != nil {
		return fmt.Errorf("unable to copy contents of the remote's file: %v", err)
	}
	return nil
}

func (t *Target) copyFileFromTarget(srcPath, dstPath string) error {
	var srcFile, dstFile File

	t.logger.Debug().
		Str("srcPath", srcPath).
		Str("dstPath", dstPath).
		Msg("Copying")

	if t.client == nil { // Local
		absSrcFilePath, err := utils.GetAbsPath(srcPath)
		if err != nil {
			return fmt.Errorf("error getting absolute src path: %w", err)
		}

		srcFile, err = os.Open(absSrcFilePath)
		if err != nil {
			return fmt.Errorf("unable to open source file: %w", err)
		}

	} else { // SSH
		var err error
		srcFile, err = t.sftp.Open(srcPath)
		if err != nil {
			return fmt.Errorf("unable to open source file: %v", err)
		}
	}

	defer srcFile.Close()

	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("unable to read metadata of source file: %v", err)
	}

	absDstPath, err := utils.GetAbsPath(dstPath)
	if err != nil {
		return fmt.Errorf("error getting absolute dst path: %w", err)
	}

	dstFile, err = utils.CreateFile(absDstPath)
	if err != nil {
		return err
	}

	defer dstFile.Close()

	dstFile.Chmod(srcFileInfo.Mode())

	if _, err = dstFile.ReadFrom(srcFile); err != nil {
		return fmt.Errorf("unable to copy contents of the remote's file: %v", err)
	}
	return nil
}

func (t *Target) Exec(ctx context.Context, env environment.Environment, name string, arg ...string) (*ExecutionResult, error) {
	if t.client == nil { // Local
		command := exec.Command(name, arg...)

		var stdout, stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr
		command.Env = os.Environ()

		for name, value := range t.environment {
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", name, value))
		}

		if env != nil {
			for name, value := range env.GetVariables() {
				command.Env = append(command.Env, fmt.Sprintf("%s=%s", t.envName(name), value))
			}
		}

		t.logger.Trace().Msgf("Environment: %s", t.environment)

		err := command.Start()
		if err != nil {
			return nil, fmt.Errorf("unable to start command locally: %s", err)
		}
		startTime := time.Now().UnixNano()

		// Do not block to enable interrupt handling and kill the process if needed
		ch := make(chan error, 1)
		go func(ch chan error) {
			ch <- command.Wait()
		}(ch)

		select {
		case err = <-ch:
			endTime := time.Now().UnixNano()
			result := ExecutionResult{
				Stdout:    stdout.Bytes(),
				Stderr:    stderr.Bytes(),
				StartTime: startTime,
				EndTime:   endTime,
			}
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					result.ExitStatus = exitErr.ExitCode()
					return &result, err
				}
				return nil, fmt.Errorf("unable to execute command locally: %s", err)
			}
			return &result, nil
		case <-ctx.Done(): // Kill process
			defer command.Process.Release()
			err = command.Process.Signal(os.Interrupt)
			if err != nil {
				err = command.Process.Signal(os.Kill)
				if err != nil {
					return nil, fmt.Errorf("killing process failed: %v", err)
				}
				return nil, fmt.Errorf("killed process")
			}
			return nil, fmt.Errorf("interrupted process")
		}
	} else { // SSH
		session, err := t.openSSHSession()
		if err != nil {
			return nil, fmt.Errorf("unable to create new ssh session: %s", err)
		}
		defer session.Close()

		var stdout, stderr bytes.Buffer
		session.Stdout = &stdout
		session.Stderr = &stderr

		for name, value := range t.environment {
			err = session.Setenv(name, value)
			if err != nil {
				return nil, fmt.Errorf("unable to set environment variable %s: %s", name, err)
			}
		}

		if env != nil {
			for name, value := range env.GetVariables() {
				err = session.Setenv(t.envName(name), value)
				if err != nil {
					return nil, fmt.Errorf("unable to set environment variable %s: %s", name, err)
				}
			}
		}

		command := fmt.Sprintf("%s %s", name, strings.Join(arg, " "))
		err = session.Start(command)
		if err != nil {
			return nil, fmt.Errorf("unable to start command on ssh server: %s", err)
		}
		startTime := time.Now().UnixNano()

		// Do not block to enable interrupt handling and kill the process if needed
		ch := make(chan error, 1)
		go func(ch chan error) {
			ch <- session.Wait()
		}(ch)

		select {
		case err = <-ch:
			endTime := time.Now().UnixNano()
			result := ExecutionResult{
				Stdout:    stdout.Bytes(),
				Stderr:    stderr.Bytes(),
				StartTime: startTime,
				EndTime:   endTime,
			}
			if err != nil {
				if exitErr, ok := err.(*ssh.ExitError); ok {
					result.ExitStatus = exitErr.ExitStatus()
					return &result, err
				}
				if _, ok := err.(*ssh.ExitMissingError); ok {
					return &result, err
				}
				return nil, fmt.Errorf("unable to execute command on ssh server: %s", err)
			}
			return &result, nil
		case <-ctx.Done(): // Kill process
			err = session.Signal(ssh.SIGINT)
			if err != nil {
				err = session.Signal(ssh.SIGKILL)
				if err != nil {
					return nil, fmt.Errorf("killing process failed: %v", err)
				}
				return nil, fmt.Errorf("killed process")
			}
			return nil, fmt.Errorf("interrupted process")
		}
	}
}
