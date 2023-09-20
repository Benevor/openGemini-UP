package start

import (
	"errors"
	"fmt"
	"openGemini-UP/pkg/config"
	"openGemini-UP/pkg/exec"
	"openGemini-UP/pkg/install"
	"openGemini-UP/pkg/status"
	"openGemini-UP/util"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type Starter interface {
	PrepareForStart() error
	Start() error
	Close()
}

type GeminiStarter struct {
	version string
	// ip -> remotes
	remotes map[string]*config.RemoteHost
	runs    *exec.RunActions

	// ip -> ssh clients
	sshClients map[string]*ssh.Client

	configurator config.Configurator // conf reader
	executor     exec.Executor       // execute commands on remote host

	clusterOptions install.ClusterOptions

	wg sync.WaitGroup
}

func NewGeminiStarter(ops install.ClusterOptions) Starter {
	return &GeminiStarter{
		remotes:        make(map[string]*config.RemoteHost),
		sshClients:     make(map[string]*ssh.Client),
		version:        ops.Version,
		configurator:   config.NewGeminiConfigurator(ops.YamlPath, filepath.Join(util.Download_dst, ops.Version, util.Local_etc_rel_path, util.Local_conf_name), filepath.Join(util.Download_dst, util.Local_etc_rel_path)),
		runs:           &exec.RunActions{},
		clusterOptions: ops,
	}
}

func (d *GeminiStarter) PrepareForStart() error {
	var err error
	if err = d.configurator.RunWithoutGen(); err != nil {
		return err
	}
	conf := d.configurator.GetConfig()

	if err = d.prepareRemotes(conf, false); err != nil {
		fmt.Printf("Failed to establish SSH connections with all remote servers. The specific error is: %s\n", err)
		return err
	}
	fmt.Println("Success to establish SSH connections with all remote servers.")

	d.executor = exec.NewGeminiExecutor(d.sshClients)

	// check process conflict and port conflict
	if d.checkProcessConflict() {
		return errors.New("process conflict before starting")
	}
	if confilted, port := d.checkPortConflict(conf); confilted {
		return fmt.Errorf("port %d conflict before starting", port)
	}

	if err = d.prepareRunActions(conf); err != nil {
		return err
	}
	return nil
}

func (d *GeminiStarter) prepareRemotes(c *config.Config, needSftp bool) error {
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

	if err := d.tryConnect(); err != nil {
		return err
	}

	return nil
}

func (d *GeminiStarter) tryConnect() error {
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
	}
	return nil
}

func (d *GeminiStarter) prepareRunActions(c *config.Config) error {
	// ts-meta
	i := 1
	for _, host := range c.CommonConfig.MetaHosts {
		d.runs.MetaAction = append(d.runs.MetaAction, &exec.RunAction{
			Info: &exec.RunInfo{
				ScriptPath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path, util.Install_Script),
				Args: []string{util.TS_META, d.remotes[host].LogPath,
					filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_bin_rel_path, util.TS_META),
					filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path, host+util.Remote_conf_suffix),
					filepath.Join(d.remotes[host].LogPath, util.Remote_pid_path, util.META+strconv.Itoa(i)+util.Remote_pid_suffix),
					filepath.Join(d.remotes[host].LogPath, strconv.Itoa(i), util.META_extra_log+strconv.Itoa(i)+util.Remote_log_suffix),
					strconv.Itoa(i)},
			},
			Remote: d.remotes[host],
		})
		i++
	}

	// ts-sql
	i = 1
	for _, host := range c.CommonConfig.SqlHosts {
		d.runs.SqlAction = append(d.runs.SqlAction, &exec.RunAction{
			Info: &exec.RunInfo{
				ScriptPath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path, util.Install_Script),
				Args: []string{util.TS_SQL, d.remotes[host].LogPath,
					filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_bin_rel_path, util.TS_SQL),
					filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path, host+util.Remote_conf_suffix),
					filepath.Join(d.remotes[host].LogPath, util.Remote_pid_path, util.SQL+strconv.Itoa(i)+util.Remote_pid_suffix),
					filepath.Join(d.remotes[host].LogPath, strconv.Itoa(i), util.SQL_extra_log+strconv.Itoa(i)+util.Remote_log_suffix),
					strconv.Itoa(i)},
			},
			Remote: d.remotes[host],
		})
		i++
	}

	// ts-store
	i = 1
	for _, host := range c.CommonConfig.StoreHosts {
		d.runs.StoreAction = append(d.runs.StoreAction, &exec.RunAction{
			Info: &exec.RunInfo{
				ScriptPath: filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path, util.Install_Script),
				Args: []string{util.TS_STORE, d.remotes[host].LogPath,
					filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_bin_rel_path, util.TS_STORE),
					filepath.Join(d.remotes[host].UpDataPath, d.version, util.Remote_etc_rel_path, host+util.Remote_conf_suffix),
					filepath.Join(d.remotes[host].LogPath, util.Remote_pid_path, util.STORE+strconv.Itoa(i)+util.Remote_pid_suffix),
					filepath.Join(d.remotes[host].LogPath, strconv.Itoa(i), util.STORE_extra_log+strconv.Itoa(i)+util.Remote_log_suffix),
					strconv.Itoa(i)},
			},
			Remote: d.remotes[host],
		})
		i++
	}

	return nil
}

func (d *GeminiStarter) checkProcessConflict() bool {
	for ip := range d.remotes {
		output, err := d.executor.ExecCommand(ip, status.CheckProcessCommand)
		if err != nil {
			fmt.Println(err)
			return true
		} else {
			if output != "" {
				return true
			}
		}
	}
	return false
}

func (d *GeminiStarter) checkPortConflict(conf *config.Config) (bool, int) {
	// check port conflict about ts-meta
	for _, ip := range conf.CommonConfig.MetaHosts {
		tomlPath := filepath.Join(util.Download_dst, util.Local_etc_rel_path, ip+util.Remote_conf_suffix)
		t, err := config.ReadFromToml(tomlPath)
		if err != nil {
			fmt.Println(err)
			return true, 0
		}

		if err, confilted, port := d.checkOnePortWithStr(ip, t.Meta.BindAddress); err != nil {
			fmt.Println(err)
			return true, 0
		} else {
			if confilted {
				return true, port
			}
		}

		if err, confilted, port := d.checkOnePortWithStr(ip, t.Meta.HttpBindAddress); err != nil {
			fmt.Println(err)
			return true, 0
		} else {
			if confilted {
				return true, port
			}
		}

		if err, confilted, port := d.checkOnePortWithStr(ip, t.Meta.RpcBindAddress); err != nil {
			fmt.Println(err)
			return true, 0
		} else {
			if confilted {
				return true, port
			}
		}

		if err, confilted, port := d.checkOnePortWithInt(ip, t.Gossip.MetaBindPort); err != nil {
			fmt.Println(err)
			return true, 0
		} else {
			if confilted {
				return true, port
			}
		}
	}

	// check port conflict about ts-sql
	for _, ip := range conf.CommonConfig.SqlHosts {
		tomlPath := filepath.Join(util.Download_dst, util.Local_etc_rel_path, ip+util.Remote_conf_suffix)
		t, err := config.ReadFromToml(tomlPath)
		if err != nil {
			fmt.Println(err)
			return true, 0
		}

		if err, confilted, port := d.checkOnePortWithStr(ip, t.Http.BindAddress); err != nil {
			fmt.Println(err)
			return true, 0
		} else {
			if confilted {
				return true, port
			}
		}
	}

	// check port conflict about ts-store
	for _, ip := range conf.CommonConfig.StoreHosts {
		tomlPath := filepath.Join(util.Download_dst, util.Local_etc_rel_path, ip+util.Remote_conf_suffix)
		t, err := config.ReadFromToml(tomlPath)
		if err != nil {
			fmt.Println(err)
			return true, 0
		}

		if err, confilted, port := d.checkOnePortWithStr(ip, t.Data.StoreIngestAddr); err != nil {
			fmt.Println(err)
			return true, 0
		} else {
			if confilted {
				return true, port
			}
		}

		if err, confilted, port := d.checkOnePortWithStr(ip, t.Data.StoreSelectAddr); err != nil {
			fmt.Println(err)
			return true, 0
		} else {
			if confilted {
				return true, port
			}
		}

		if err, confilted, port := d.checkOnePortWithInt(ip, t.Gossip.StoreBindPort); err != nil {
			fmt.Println(err)
			return true, 0
		} else {
			if confilted {
				return true, port
			}
		}
	}

	return false, 0
}

func (d *GeminiStarter) checkOnePortWithStr(ip, inputStr string) (error, bool, int) {
	parts := strings.Split(inputStr, ":")
	if len(parts) < 2 {
		return errors.New("invalid inputStr format when check one port conflit"), true, 0
	}
	portStr := parts[1]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err, true, port
	}

	output, err := d.executor.ExecCommand(ip, status.GenCheckPortCommand(port))
	if err != nil {
		return err, true, port
	} else {
		if output == "yes\n" {
			return nil, true, port
		} else if output == "no\n" {
			return nil, false, port
		}
	}
	return errors.New("unexpected port confilt check result"), true, port
}

func (d *GeminiStarter) checkOnePortWithInt(ip string, port int) (error, bool, int) {
	output, err := d.executor.ExecCommand(ip, status.GenCheckPortCommand(port))
	if err != nil {
		return err, true, port
	} else {
		if output == "yes\n" {
			return nil, true, port
		} else if output == "no\n" {
			return nil, false, port
		}
	}
	return errors.New("unexpected port confilt check result"), true, port
}

func (d *GeminiStarter) Start() error {
	if d.executor == nil {
		return nil
	}
	errChan := make(chan error, len(d.runs.MetaAction)+len(d.runs.SqlAction)+len(d.runs.StoreAction))
	// start all ts-meta concurrently
	d.wg.Add(len(d.runs.MetaAction))
	for _, action := range d.runs.MetaAction {
		go func(action *exec.RunAction, errChan chan error) {
			defer d.wg.Done()
			d.executor.ExecRunAction(action, errChan)
		}(action, errChan)
	}
	d.wg.Wait()

	// time for ts-meta campaign
	time.Sleep(time.Second)

	// start all ts-store and ts-sql concurrently
	d.wg.Add(len(d.runs.SqlAction) + len(d.runs.StoreAction))
	for _, action := range d.runs.StoreAction {
		go func(action *exec.RunAction, errChan chan error) {
			defer d.wg.Done()
			d.executor.ExecRunAction(action, errChan)
		}(action, errChan)
	}
	for _, action := range d.runs.SqlAction {
		go func(action *exec.RunAction, errChan chan error) {
			defer d.wg.Done()
			d.executor.ExecRunAction(action, errChan)
		}(action, errChan)
	}
	d.wg.Wait()

	select {
	case <-errChan:
		return errors.New("cluster start failed")
	default:
		return nil
	}
}

func (d *GeminiStarter) Close() {
	var err error
	for _, ssh := range d.sshClients {
		if err = ssh.Close(); err != nil {
			fmt.Println(err)
		}
	}
}
