package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Toml struct {
	Common  CommonToml  `toml:"common"`
	Meta    MetaToml    `toml:"meta"`
	Http    HttpToml    `toml:"http"`
	Data    DataToml    `toml:"data"`
	Logging LoggingToml `toml:"logging"`
	Gossip  GossipToml  `toml:"gossip"`
}

type CommonToml struct {
	MetaJoin []string `toml:"meta-join"`
}

type MetaToml struct {
	BindAddress     string `toml:"bind-address"`
	HttpBindAddress string `toml:"http-bind-address"`
	RpcBindAddress  string `toml:"rpc-bind-address"`
	Dir             string `toml:"dir"`
}

type HttpToml struct {
	BindAddress string `toml:"bind-address"`
}

type DataToml struct {
	StoreIngestAddr     string `toml:"store-ingest-addr"`
	StoreSelectAddr     string `toml:"store-select-addr"`
	StoreDataDir        string `toml:"store-data-dir"`
	StoreWalDir         string `toml:"store-wal-dir"`
	StoreMetaDir        string `toml:"store-meta-dir"`
	CacheTableDataBlock bool   `toml:"cache-table-data-block"`
	CacheTableMetaBlock bool   `toml:"cache-table-meta-block"`
	ReadCacheLimit      int    `toml:"read-cache-limit"`
}

type LoggingToml struct {
	Path string `toml:"path"`
}

type GossipToml struct {
	BindAddress   string   `toml:"bind-address"`
	StoreBindPort int      `toml:"store-bind-port"`
	MetaBindPort  int      `toml:"meta-bind-port"`
	Members       []string `toml:"members"`
}

func ReadFromToml(tomlPath string) (Toml, error) {
	t := Toml{}
	var err error
	if _, err = toml.DecodeFile(tomlPath, &t); err != nil {
		return t, err
	}
	return t, nil
}

func ConvertToml(hostToml Toml, pwd string) Toml {
	if len(hostToml.Meta.Dir) > 1 && hostToml.Meta.Dir[:1] == "~" {
		hostToml.Meta.Dir = filepath.Join(pwd, hostToml.Meta.Dir[1:])
	}
	if len(hostToml.Data.StoreDataDir) > 1 && hostToml.Data.StoreDataDir[:1] == "~" {
		hostToml.Data.StoreDataDir = filepath.Join(pwd, hostToml.Data.StoreDataDir[1:])
	}
	if len(hostToml.Data.StoreWalDir) > 1 && hostToml.Data.StoreWalDir[:1] == "~" {
		hostToml.Data.StoreWalDir = filepath.Join(pwd, hostToml.Data.StoreWalDir[1:])
	}
	if len(hostToml.Data.StoreMetaDir) > 1 && hostToml.Data.StoreMetaDir[:1] == "~" {
		hostToml.Data.StoreMetaDir = filepath.Join(pwd, hostToml.Data.StoreMetaDir[1:])
	}
	if len(hostToml.Logging.Path) > 1 && hostToml.Logging.Path[:1] == "~" {
		hostToml.Logging.Path = filepath.Join(pwd, hostToml.Logging.Path[1:])
	}
	return hostToml
}

func GenNewToml(t Toml, path string) {
	f, _ := os.Create(path)
	defer f.Close()

	e := toml.NewEncoder(f)
	e.Encode(t)
}
