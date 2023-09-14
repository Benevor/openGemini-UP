package util

import "time"

// TODO:这里的大多数后面要改成从配置文件中读取，这里仅仅存储一个默认值

// downloader
const (
	Download_web       = "https://github.com/openGemini/openGemini/releases/download"
	Download_version   = "v1.0.0"
	Download_fill_char = "openGemini-"
	Download_type      = "-linux-amd64.tar.gz"
	Download_dst       = "./data"
	Download_timeout   = 2 * time.Minute
)

// local
const (
	Local_bin_rel_path = "usr/bin"
	Local_etc_rel_path = "etc"
	Local_conf_name    = "openGemini.conf"
)

// config
const (
	User_conf_path      = "./topology.example.yaml"
	Install_script_path = "./scripts/install.sh"
	Remote_conf_suffix  = "-openGemini.conf"
	SSH_KEY             = "SSH_KEY"
	SSH_PW              = "SSH_PW"
)

// file name
const (
	TS_META        = "ts-meta" // process name & bin file name
	TS_SQL         = "ts-sql"
	TS_STORE       = "ts-store"
	Install_Script = "install.sh"
)

// remote
const (
	// openGemini-UP
	Remote_bin_rel_path = "bin"
	Remote_etc_rel_path = "etc"

	// openGemini
	OpenGemini_path   = "/tmp/openGemini"
	Remote_pid_path   = "pid"
	Remote_log_path   = "logs"
	Remote_pid_suffix = ".pid"
	Remote_log_suffix = ".log"

	META_extra_log  = "meta_extra"
	SQL_extra_log   = "sql_extra"
	STORE_extra_log = "store_extra"
	META            = "meta"
	SQL             = "sql"
	STORE           = "store"
)

// version
const (
	VersionFile = "version"
)

// yaml default values
// const (
// 	// ts-meta
// 	Meta_client_port = 8091
// 	Meta_peer_port   = 8092
// 	Meta_raft_port   = 8088
// 	Meta_gossip_port = 8010
// 	Meta_data_dir    = "/gemini-data/meta"

// 	// te-sql
// 	Sql_port        = 8086
// 	Sql_flight_port = 8087

// 	// ts-store
// 	Store_ingest_port    = 8400
// 	Store_select_port    = 8401
// 	Store_gossip_port    = 8011
// 	Store_store_data_dir = "/gemini-data/data"
// 	Store_meta_dir       = "/gemini-data/data/meta"
// )
