package target

import (
	"errors"
	"fmt"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/cmd/common"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/utils"
	"golang.org/x/crypto/ssh"
	"io"
	"io/fs"
	"os"
)

type File interface {
	io.Reader

	Close() error
	Chmod(fs.FileMode) error
	ReadFrom(io.Reader) (int64, error)
	Stat() (os.FileInfo, error)
}

func connectSSH(config *common.SSHConfig) (*ssh.Client, error) {
	signer, err := loadSSHKey(config)
	if err != nil {
		return nil, fmt.Errorf("unable to load ssh key: %s", err)
	}

	//absoluteKnownHostsPath, err := utils.GetAbsPath(config.Knownhosts)
	//if err != nil {
	//	return nil, fmt.Errorf("error getting absolute knownhosts path: %w", err)
	//}
	//
	//hostKeys, err := knownhosts.New(absoluteKnownHostsPath)
	//if err != nil {
	//	return nil, fmt.Errorf("unable to read ssh's \"known_hosts\" file: %s", err)
	//}

	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.Hostname, config.Port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to ssh server: %s", err)
	}
	return sshClient, nil
}

func loadSSHKey(config *common.SSHConfig) (ssh.Signer, error) {
	absolutePrivateKeyPath, err := utils.GetAbsPath(config.Privkey)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute file path of private key file: %w", err)
	}

	key, err := os.ReadFile(absolutePrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read the ssh private key: %s", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		var keyErr *ssh.PassphraseMissingError
		if errors.As(err, &keyErr) {
			password, err := utils.ReadSecretFromStdin(
				fmt.Sprintf("Password for SSH key (%s): ", config.Hostname),
			)
			if err != nil {
				return nil, fmt.Errorf("error during password read: %v", err)
			}
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, password)
			if err != nil {
				return nil, fmt.Errorf("unable to parse the ssh private key: %s", err)
			}
		} else {
			return nil, fmt.Errorf("unable to parse the ssh private key: %s", err)
		}
	}
	return signer, nil
}
