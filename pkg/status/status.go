package status

import (
	"errors"
	"fmt"
	"openGemini-UP/pkg/config"
	"openGemini-UP/pkg/exec"
	"openGemini-UP/pkg/install"
	"openGemini-UP/util"
	"os"
	"strings"
	"sync"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh"
)

const CheckProcessCommand = "ps aux | grep -E '(ts-meta|ts-sql|ts-store)' | grep -v grep | awk '{print $11}'"
const CheckDiskCapicityCommand = "df -h | grep '^/dev/'"

var AlwaysCheckPosrt = []int{8088, 8091, 8092, 8086, 8400, 8401, 8010, 8011}

func GenCheckPortCommand(port int) string {
	return fmt.Sprintf("ss -tln | grep -q ':%d' && echo 'yes' || echo 'no'", port)
}

type ClusterStatusPerServer struct {
	Ip               string
	RunningProcesses []string     // ts-meta,ts-sql,ts-store
	PortOccupancy    map[int]bool // port->occupancy or not
	DiskCapacity     []string     // disk->capacity
}

type StatusPatroller interface {
	PrepareForPatrol() error
	Patrol() error
	Close()
}

type GeminiStatusPatroller struct {
	// ip -> remotes
	remotes map[string]*config.RemoteHost

	// ip -> ssh clients
	sshClients map[string]*ssh.Client

	configurator config.Configurator // conf reader
	executor     exec.Executor       // execute commands on remote host

	clusterOptions install.ClusterOptions

	wg sync.WaitGroup
}

func NewGeminiStatusPatroller(ops install.ClusterOptions) StatusPatroller {
	return &GeminiStatusPatroller{
		remotes:        make(map[string]*config.RemoteHost),
		sshClients:     make(map[string]*ssh.Client),
		configurator:   config.NewGeminiConfigurator(ops.YamlPath, "", ""),
		clusterOptions: ops,
	}
}

func (d *GeminiStatusPatroller) PrepareForPatrol() error {
	var err error
	if err = d.configurator.RunWithoutGen(); err != nil {
		return err
	}
	conf := d.configurator.GetConfig()

	// check the internet with all the remote servers
	if err = d.prepareRemotes(conf); err != nil {
		fmt.Printf("Failed to establish SSH connections with all remote servers. The specific error is: %s\n", err)
		return err
	}
	fmt.Println("Success to establish SSH connections with all remote servers.")

	d.executor = exec.NewGeminiExecutor(d.sshClients)
	return nil
}

func (d *GeminiStatusPatroller) prepareRemotes(c *config.Config) error {
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

func (d *GeminiStatusPatroller) tryConnect() error {
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

func (d *GeminiStatusPatroller) Patrol() error {
	statusChan := make(chan ClusterStatusPerServer, len(d.remotes))
	errChan := make(chan error, len(d.remotes))
	d.wg.Add(len(d.remotes))
	for ip := range d.remotes {
		go d.patrolOneServer(ip, statusChan, errChan)
	}
	d.wg.Wait()

	select {
	case <-errChan:
		close(errChan)
		return errors.New("check cluster status failed")
	default:
	}

	for {
		select {
		case status := <-statusChan:
			displayGeminiStatus(status)
		default:
			return nil
		}
	}
}

func (d *GeminiStatusPatroller) patrolOneServer(ip string, statusChan chan ClusterStatusPerServer, errChan chan error) {
	defer d.wg.Done()
	var status = ClusterStatusPerServer{
		Ip:            ip,
		PortOccupancy: make(map[int]bool),
	}
	var err error

	// check running process
	output, err := d.executor.ExecCommand(ip, CheckProcessCommand)
	if err != nil {
		fmt.Println(err)
		errChan <- err
		return
	} else {
		status.RunningProcesses = strings.Split(output, "\n")
	}

	// check port occupancy
	y, err := config.ReadFromYaml(d.clusterOptions.YamlPath)
	if err != nil {
		fmt.Println(err)
		errChan <- err
		return
	}
	ports := make(map[int]int)
	for _, meta := range y.TsMeta {
		if meta.Host == ip {
			if meta.ClientPort != 0 {
				ports[meta.ClientPort] = meta.ClientPort
			}
			if meta.PeerPort != 0 {
				ports[meta.PeerPort] = meta.PeerPort
			}
			if meta.RaftPort != 0 {
				ports[meta.RaftPort] = meta.RaftPort
			}
			if meta.GossipPort != 0 {
				ports[meta.GossipPort] = meta.GossipPort
			}
		}
	}
	for _, sql := range y.TsSql {
		if sql.Host == ip {
			if sql.Port != 0 {
				ports[sql.Port] = sql.Port
			}
			if sql.FlightPort != 0 {
				ports[sql.FlightPort] = sql.FlightPort
			}
		}
	}
	for _, store := range y.TsStore {
		if store.Host == ip {
			if store.IngestPort != 0 {
				ports[store.IngestPort] = store.IngestPort
			}
			if store.SelectPort != 0 {
				ports[store.SelectPort] = store.SelectPort
			}
			if store.GossipPort != 0 {
				ports[store.GossipPort] = store.GossipPort
			}
		}
	}
	for _, p := range AlwaysCheckPosrt {
		ports[p] = p
	}
	for port := range ports {
		output, err := d.executor.ExecCommand(ip, GenCheckPortCommand(port))
		if err != nil {
			fmt.Println(err)
			errChan <- err
			return
		} else {
			if output == "yes\n" {
				status.PortOccupancy[port] = true
			} else if output == "no\n" {
				status.PortOccupancy[port] = false
			} else {
				fmt.Printf("unexpected output when checking port %d. %s\n", port, output)
				errChan <- errors.New("unexpected output when checking port")
				return
			}
		}
	}

	// check disk capacity
	output, err = d.executor.ExecCommand(ip, CheckDiskCapicityCommand)
	if err != nil {
		fmt.Println(err)
		errChan <- err
		return
	} else {
		status.DiskCapacity = strings.Split(output, "\n")
	}
	statusChan <- status
}

func (d *GeminiStatusPatroller) Close() {
	var err error
	for _, ssh := range d.sshClients {
		if err = ssh.Close(); err != nil {
			fmt.Println(err)
		}
	}
}

func displayGeminiStatus(status ClusterStatusPerServer) {
	fmt.Printf("\nGemini status of server %s\n", status.Ip)

	// Create a new table for Running Processes
	runningProcessesTable := tablewriter.NewWriter(os.Stdout)
	runningProcessesTable.SetHeader([]string{"Running Processes"})
	for _, process := range status.RunningProcesses {
		runningProcessesTable.Append([]string{fmt.Sprintf("%v", process)})
	}
	runningProcessesTable.Render()

	// Create a new table for Port Occupancy
	portOccupancyTable := tablewriter.NewWriter(os.Stdout)
	portOccupancyTable.SetHeader([]string{"Port Occupancy"})
	for port, occupied := range status.PortOccupancy {
		portOccupancyTable.Append([]string{fmt.Sprintf("Port %d: Occupied: %v", port, occupied)})
	}
	portOccupancyTable.Render()

	// Create a new table for Disk Capacity
	diskCapacityTable := tablewriter.NewWriter(os.Stdout)
	diskCapacityTable.SetHeader([]string{"Disk Capacity"})
	for _, capacity := range status.DiskCapacity {
		diskCapacityTable.Append([]string{fmt.Sprintf("%v", capacity)})
	}
	diskCapacityTable.Render()
}
