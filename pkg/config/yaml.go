package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Yaml struct {
	Global  GlobalYaml  `yaml:"global"`
	TsMeta  []MetaYaml  `yaml:"ts-meta"`
	TsSql   []SqlYaml   `yaml:"ts-sql"`
	TsStore []StoreYaml `yaml:"ts-store"`
}

type GlobalYaml struct {
	SSHPort   int    `yaml:"ssh_port"`
	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`

	User        string `yaml:"user"`
	Group       string `yaml:"group"`
	BaseDataDir string `yaml:"base_data_dir"`
	OS          string `yaml:"os"`
	Arch        string `yaml:"arch"`
}

type MetaYaml struct {
	Host string `yaml:"host"`
	// default value in GlobalYaml
	SSHPort   int    `yaml:"ssh_port"`
	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`

	// default value in const
	ClientPort int    `yaml:"client_port"`
	PeerPort   int    `yaml:"peer_port"`
	RaftPort   int    `yaml:"raft_port"`
	GossipPort int    `yaml:"gossip_port"`
	DataDir    string `yaml:"data_dir"`
}

type SqlYaml struct {
	Host string `yaml:"host"`
	// default value in GlobalYaml
	SSHPort   int    `yaml:"ssh_port"`
	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`

	// default value in const
	Port       int `yaml:"port"`
	FlightPort int `yaml:"flight_port"`
}

type StoreYaml struct {
	Host string `yaml:"host"`
	// default value in GlobalYaml
	SSHPort   int    `yaml:"ssh_port"`
	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`

	// default value in const
	IngestPort int    `yaml:"ingest_port"`
	SelectPort int    `yaml:"select_port"`
	GossipPort int    `yaml:"gossip_port"`
	DataDir    string `yaml:"data_dir"`
	MetaDir    string `yaml:"meta_dir"`
}

func ReadFromYaml(yamlPath string) (Yaml, error) {
	var err error
	yamlFile, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return Yaml{}, err
	}
	var y Yaml
	if err = yaml.Unmarshal(yamlFile, &y); err != nil {
		return Yaml{}, err
	}
	return y, nil
}
