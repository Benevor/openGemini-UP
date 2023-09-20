package uninstall

import (
	"fmt"
	"openGemini-UP/pkg/config"
	"openGemini-UP/pkg/exec"
	"openGemini-UP/pkg/install"
	"openGemini-UP/util"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/ssh"
)

type Uninstall interface {
	Prepare() error
	Run() error
	Close()
}

type GeminiUninstaller struct {
	// ip -> remotes
	remotes map[string]*config.RemoteHost
	// ip -> ssh clients
	sshClients map[string]*ssh.Client

	configurator config.Configurator // conf reader
	executor     exec.Executor       // execute commands on remote host
	upDataPath   map[string]string   // ip->up path

	wg sync.WaitGroup

	clusterOptions install.ClusterOptions
}

func NewGeminiUninstaller(ops install.ClusterOptions) Uninstall {
	new := &GeminiUninstaller{
		remotes:        make(map[string]*config.RemoteHost),
		sshClients:     make(map[string]*ssh.Client),
		configurator:   config.NewGeminiConfigurator(ops.YamlPath, "", ""),
		upDataPath:     make(map[string]string),
		clusterOptions: ops,
	}
	return new
}

func (s *GeminiUninstaller) Prepare() error {
	var err error
	if err = s.configurator.RunWithoutGen(); err != nil {
		return err
	}
	conf := s.configurator.GetConfig()

	if err = s.prepareRemotes(conf); err != nil {
		fmt.Printf("Failed to establish SSH connections with all remote servers. The specific error is: %s\n", err)
		return err
	}
	fmt.Println("Success to establish SSH connections with all remote servers.")

	s.executor = exec.NewGeminiExecutor(s.sshClients)

	return nil
}

func (s *GeminiUninstaller) prepareRemotes(c *config.Config) error {
	if c == nil {
		return util.UnexpectedNil
	}

	for ip, ssh := range c.SSHConfig {
		s.remotes[ip] = &config.RemoteHost{
			Ip:       ip,
			SSHPort:  ssh.Port,
			User:     s.clusterOptions.User,
			Password: s.clusterOptions.Password,
			KeyPath:  s.clusterOptions.Key,
			Typ:      s.clusterOptions.SshType,
		}

		s.upDataPath[ip] = ssh.UpDataPath
	}

	if err := s.tryConnect(); err != nil {
		return err
	}

	return nil
}

func (s *GeminiUninstaller) tryConnect() error {
	for ip, r := range s.remotes {
		var err error
		var sshClient *ssh.Client
		switch r.Typ {
		case config.SSH_PW:
			sshClient, err = util.NewSSH_PW(r.User, r.Password, r.Ip, r.SSHPort)
		case config.SSH_KEY:
			sshClient, err = util.NewSSH_Key(r.User, r.KeyPath, r.Ip, r.SSHPort)

		}
		if err != nil {
			// TODO(Benevor):close all connection and exit
			return err
		}
		s.sshClients[ip] = sshClient
	}
	return nil
}

func (s *GeminiUninstaller) Run() error {
	if s.executor == nil {
		return util.UnexpectedNil
	}
	s.wg.Add(len(s.remotes))
	for ip := range s.remotes {
		go func(ip string) {
			defer s.wg.Done()
			command := fmt.Sprintf("rm -rf %s;", filepath.Join(s.upDataPath[ip], s.clusterOptions.Version))
			s.executor.ExecCommand(ip, command)
		}(ip)
	}
	s.wg.Wait()
	return nil
}

func (s *GeminiUninstaller) Close() {
	var err error
	for _, ssh := range s.sshClients {
		if ssh != nil {
			if err = ssh.Close(); err != nil {
				fmt.Println(err)
			}
		}
	}
}
