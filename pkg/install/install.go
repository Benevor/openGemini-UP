package install

import (
	"errors"
	"fmt"
	"openGemini-UP/pkg/config"
	"openGemini-UP/pkg/download"
	"openGemini-UP/pkg/exec"
	"openGemini-UP/util"
	"path/filepath"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type ClusterOptions struct {
	Version  string
	User     string
	Key      string
	Password string
	SshType  config.SSHType
	YamlPath string
}

type UploadAction struct {
	uploadInfo []*config.UploadInfo
	remoteHost *config.RemoteHost
}

type Installer interface {
	PrepareForInstall() error
	Install() error
	Close()
}

type GeminiInstaller struct {
	version string
	// ip -> remotes
	remotes map[string]*config.RemoteHost
	uploads map[string]*UploadAction

	// ip -> ssh clients
	sshClients  map[string]*ssh.Client
	sftpClients map[string]*sftp.Client

	configurator config.Configurator // conf reader
	executor     exec.Executor       // execute commands on remote host

	clusterOptions ClusterOptions

	wg sync.WaitGroup
}

func NewGeminiInstaller(ops ClusterOptions) Installer {
	return &GeminiInstaller{
		remotes:        make(map[string]*config.RemoteHost),
		uploads:        make(map[string]*UploadAction),
		sshClients:     make(map[string]*ssh.Client),
		sftpClients:    make(map[string]*sftp.Client),
		version:        ops.Version,
		configurator:   config.NewGeminiConfigurator(ops.YamlPath, filepath.Join(util.Download_dst, ops.Version, util.Local_etc_rel_path, util.Local_conf_name), filepath.Join(util.Download_dst, util.Local_etc_rel_path)),
		clusterOptions: ops,
	}
}

func (d *GeminiInstaller) PrepareForInstall() error {
	var err error
	if err = d.configurator.Run(); err != nil {
		return err
	}
	conf := d.configurator.GetConfig()

	dOps := download.DownloadOptions{
		Version: d.version,
		Os:      conf.CommonConfig.Os,
		Arch:    conf.CommonConfig.Arch,
	}
	downloader := download.NewGeminiDownloader(dOps)
	if err = downloader.Run(); err != nil {
		return err
	}

	// check the internet with all the remote servers
	if err = d.prepareRemotes(conf, true); err != nil {
		fmt.Printf("Failed to establish SSH connections with all remote servers. The specific error is: %s\n", err)
		return err
	}
	fmt.Println("Success to establish SSH connections with all remote servers.")

	d.executor = exec.NewGeminiExecutor(d.sshClients)

	if err = d.prepareForUpload(); err != nil {
		return err
	}

	if err = d.prepareUploadActions(conf); err != nil {
		return err
	}

	return nil
}

func (d *GeminiInstaller) prepareRemotes(c *config.Config, needSftp bool) error {
	if c == nil {
		return util.UnexpectedNil
	}

	for ip, ssh := range c.SSHConfig {
		d.remotes[ip] = &config.RemoteHost{
			Ip:         ip,
			SSHPort:    ssh.Port,
			UpDataPath: ssh.UpDataPath,
			LogPath:    ssh.LogPath,
			User:       d.clusterOptions.User,
			Typ:        d.clusterOptions.SshType,
			Password:   d.clusterOptions.Password,
			KeyPath:    d.clusterOptions.Key,
		}
	}

	if err := d.tryConnect(needSftp); err != nil {
		return err
	}

	return nil
}

func (d *GeminiInstaller) tryConnect(needSftp bool) error {
	for ip, r := range d.remotes {
		var err error
		var sshClient *ssh.Client
		switch r.Typ {
		case config.SSH_PW:
			sshClient, err = util.NewSSH_PW(r.User, r.Password, r.Ip, r.SSHPort)
		case config.SSH_KEY:
			sshClient, err = util.NewSSH_Key(r.User, r.KeyPath, r.Ip, r.SSHPort)

		}
		if err != nil {
			return err
		}
		d.sshClients[ip] = sshClient

		if needSftp {
			sftpClient, err := util.NewSftpClient(sshClient)
			if err != nil {
				return err
			}
			d.sftpClients[ip] = sftpClient

			pwd, _ := sftpClient.Getwd()
			// Convert relative paths to absolute paths.
			if len(r.UpDataPath) > 1 && r.UpDataPath[:1] == "~" {
				r.UpDataPath = filepath.Join(pwd, r.UpDataPath[1:])
			}

			confPath := filepath.Join(util.Download_dst, util.Local_etc_rel_path, r.Ip+util.Remote_conf_suffix)
			hostToml, _ := config.ReadFromToml(confPath)
			// Convert relative paths in openGemini.conf to absolute paths.
			hostToml = config.ConvertToml(hostToml, pwd)
			config.GenNewToml(hostToml, confPath)
		}
	}
	return nil
}

func (d *GeminiInstaller) prepareForUpload() error {
	if d.executor == nil {
		return util.UnexpectedNil
	}
	for ip, r := range d.remotes {
		binPath := filepath.Join(r.UpDataPath, d.version, util.Remote_bin_rel_path)
		etcPath := filepath.Join(r.UpDataPath, d.version, util.Remote_etc_rel_path)
		command := fmt.Sprintf("mkdir -p %s; mkdir -p %s;", binPath, etcPath)
		if _, err := d.executor.ExecCommand(ip, command); err != nil {
			return err
		}
	}
	return nil
}

func (d *GeminiInstaller) prepareUploadActions(c *config.Config) error {
	// ts-meta
	for _, host := range c.CommonConfig.MetaHosts {
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  filepath.Join(util.Download_dst, d.version, util.Local_bin_rel_path, util.TS_META),
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_bin_rel_path),
			FileName:   util.TS_META,
		})
	}

	// ts-sql
	for _, host := range c.CommonConfig.SqlHosts {
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  filepath.Join(util.Download_dst, d.version, util.Local_bin_rel_path, util.TS_SQL),
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_bin_rel_path),
			FileName:   util.TS_SQL,
		})
	}

	// ts-store
	for _, host := range c.CommonConfig.StoreHosts {
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  filepath.Join(util.Download_dst, d.version, util.Local_bin_rel_path, util.TS_STORE),
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_bin_rel_path),
			FileName:   util.TS_STORE,
		})
	}

	// conf and script
	for host := range c.SSHConfig {
		if d.uploads[host] == nil {
			d.uploads[host] = &UploadAction{
				remoteHost: d.remotes[host],
			}
		}
		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  filepath.Join(util.Download_dst, util.Local_etc_rel_path, host+util.Remote_conf_suffix),
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path),
			FileName:   host + util.Remote_conf_suffix,
		})

		d.uploads[host].uploadInfo = append(d.uploads[host].uploadInfo, &config.UploadInfo{
			LocalPath:  util.Install_script_path,
			RemotePath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path),
			FileName:   util.Install_Script,
		})
	}

	return nil
}

func (d *GeminiInstaller) Install() error {
	fmt.Println("Start to install openGemini...")
	errChan := make(chan error, len(d.uploads))
	var wgp sync.WaitGroup
	wgp.Add(2)

	go func() {
		defer wgp.Done()
		d.wg.Add(len(d.uploads))
		for ip, action := range d.uploads {
			go func(ip string, action *UploadAction, errChan chan error) {
				defer d.wg.Done()
				for _, c := range action.uploadInfo {
					// check whether need to upload the file
					// only support Linux
					cmd := fmt.Sprintf("if [ -f %s ]; then echo 'File exists'; else echo 'File not found'; fi", filepath.Join(c.RemotePath, c.FileName))
					output, err := d.executor.ExecCommand(ip, cmd)
					if string(output) == "File exists\n" && err == nil {
						fmt.Printf("%s exists on %s.\n", c.FileName, c.RemotePath)
					} else {
						if err := util.UploadFile(action.remoteHost.Ip, c.LocalPath, c.RemotePath, d.sftpClients[action.remoteHost.Ip]); err != nil {
							fmt.Printf("upload %s to %s error: %v\n", c.LocalPath, action.remoteHost.Ip, err)
							errChan <- err
						}
					}
				}

			}(ip, action, errChan)
		}
		d.wg.Wait()
		close(errChan)
	}()

	var has_err = false
	go func() {
		defer wgp.Done()
		for {
			err, ok := <-errChan
			if !ok {
				break
			}
			fmt.Println(err)
			has_err = true
		}
	}()

	wgp.Wait()
	if has_err {
		return errors.New("install cluster failed")
	} else {
		return nil
	}
}

func (d *GeminiInstaller) Close() {
	var err error
	for _, sftp := range d.sftpClients {
		if sftp != nil {
			if err = sftp.Close(); err != nil {
				fmt.Println(err)
			}
		}
	}

	for _, ssh := range d.sshClients {
		if err = ssh.Close(); err != nil {
			fmt.Println(err)
		}
	}
}
