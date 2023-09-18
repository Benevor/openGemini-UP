# openGemini-UP

Cluster deployment and upgrade tool.

## Build

```bash
git clone git@github.com:openGemini/openGemini-UP.git

cd openGemini-UP

go mod tidy

go build
```

## Commands

The following table describes some commonly used basic commands.

| command | description | parameter | example |
| --- | --- | --- | --- |
| `version` | display version number of openGemini-UP | no para | `./openGemini-UP version` |
| `list` | display the version information of all components currently downloaded | no para | `./openGemini-UP list` |
| `install` | install database components | --version | `./openGemini-UP install --version v1.0.0` |
| `cluster` | deploying and managing openGemini clusters | have subcommand | |

The following table describes the subcommands of the `cluster` command.

| command | description | parameter | example |
| --- | --- | --- | --- |
| `deploy` | deploy an openGemini cluster| --version<br />--yaml<br />--user<br />--identity_file<br />--password | `./openGemini-UP cluster deploy --version v1.0.0 --yaml ./topology.example.yaml --user root --identity_file ~/.ssh/id_rsa` |
| `stop` | stop an openGemini cluster | --yaml<br />--user<br />--identity_file<br />--password | `./openGemini-UP cluster stop --yaml ./topology.example.yaml --user root --password xxxxxx` |
| `start` | start an openGemini cluster which is stopped | --yaml<br />--user<br />--identity_file<br />--password | `./openGemini-UP cluster start --yaml ./topology.example.yaml --user root --password xxxxxx` |
| `destroy` | destroy an openGemini cluster which means stopping services and clearing data| --yaml<br />--user<br />--identity_file<br />--password | `./openGemini-UP cluster destroy --yaml ./topology.example.yaml --user root --password xxxxxx` |
| `upgrade` | upgrade an openGemini cluster to the specified version | --version<br />--yaml<br />--user<br />--identity_file<br />--password | `./openGemini-UP cluster upgrade --version v1.0.0 --yaml ./topology.example.yaml --user root --password xxxxxx` |

## topology.example.yaml

The `topology.example.yaml` is written by the user and contains the necessary information for deploying the openGemini cluster. You can modify the content of the file according to the template.

The meaning of each part is as follows:

* `global`: Default values for some options.
* `ts-meta`:  Deployment information for `ts-meta`, users can modify some options in `openGemini.conf` here.
* `ts-sql`: Deployment information for `ts-sql`, users can modify some options in `openGemini.conf` here.
* `ts-store`: Deployment information for `ts-store`, users can modify some options in `openGemini.conf` here.

```toml
# Global variables are applied to all deployments and used as the default value of
# the deployments if a specific deployment value is missing.
global:
  # The user who runs the openGemini cluster.
  user: "gemini"
  # group is used to specify the group name the user belong to,if it's not the same as user.
  group: "gemini"
  # SSH port of servers in the managed cluster.
  ssh_port: 22
  # openGemini Cluster data storage directory.
  base_data_dir: "/gemini-data"
  # openGemini Cluster log file storage directory.
  log_dir: "/gemini-deploy/logs"
  # Storage directory for cluster deployment files, startup scripts, and configuration files.
  deploy_dir: "/gemini-deploy"
  # operating system, linux/darwin/windows.
  os: "linux" 
  # Supported values: "amd64", "arm64" (default: "amd64").
  arch: "amd64"

# Server configs are used to specify the configuration of ts-meta Servers.
ts-meta:
  # The ip address of the ts-meta Server.
  - host: 10.0.1.11
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [meta].http-bind-address in openGemini.conf.
    client_port: 8091
    # [meta].rpc-bind-address in openGemini.conf.
    peer_port: 8092
    # [meta].bind-address in openGemini.conf.
    raft_port: 8088
    # [gossip].meta-bind-port in openGemini.conf.
    gossip_port: 8010
    # [meta].dir in openGemini.conf.
    data_dir: "/gemini-data/meta"
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"
  - host: 10.0.1.12
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [meta].http-bind-address in openGemini.conf.
    client_port: 8091
    # [meta].rpc-bind-address in openGemini.conf.
    peer_port: 8092
    # [meta].bind-address in openGemini.conf.
    raft_port: 8088
    # [gossip].meta-bind-port in openGemini.conf.
    gossip_port: 8010
    # [meta].dir in openGemini.conf.
    data_dir: "/gemini-data/meta"
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"
  - host: 10.0.1.13
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [meta].http-bind-address in openGemini.conf.
    client_port: 8091
    # [meta].rpc-bind-address in openGemini.conf.
    peer_port: 8092
    # [meta].bind-address in openGemini.conf.
    raft_port: 8088
    # [gossip].meta-bind-port in openGemini.conf.
    gossip_port: 8010
    # [meta].dir in openGemini.conf.
    data_dir: "/gemini-data/meta"
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"

# Server configs are used to specify the configuration of ts-sql Servers.
ts-sql:
  # The ip address of the ts-sql Server.
  - host: 10.0.1.14
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [http].bind-address in openGemini.conf.
    port: 8086
    # [http].flight-address in openGemini.conf.
    flight_port: 8087
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"
  - host: 10.0.1.15
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [http].bind-address in openGemini.conf.
    port: 8086
    # [http].flight-address in openGemini.conf.
    flight_port: 8087
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"
  - host: 10.0.1.16
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [http].bind-address in openGemini.conf.
    port: 8086
    # [http].flight-address in openGemini.conf.
    flight_port: 8087
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"

# Server configs are used to specify the configuration of ts-store Servers.
ts-store:
  # The ip address of the ts-store Server.
  - host: 10.0.1.17
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [data].store-ingest-addr in openGemini.conf.
    ingest_port: 8400
    # [data].store-select-addr in openGemini.conf.
    select_port: 8401
    # [gossip].store-bind-port in openGemini.conf.
    gossip_port: 8011
    # [data].store-data-dir & [data].store-wal-dir in openGemini.conf.
    data_dir: "/gemini-data/data"
    # [data].store-meta-dir in openGemini.conf.
    meta_dir: "/gemini-data/data/meta"
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"
  - host: 10.0.1.18
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [data].store-ingest-addr in openGemini.conf.
    ingest_port: 8400
    # [data].store-select-addr in openGemini.conf.
    select_port: 8401
    # [gossip].store-bind-port in openGemini.conf.
    gossip_port: 8011
    # [data].store-data-dir & [data].store-wal-dir in openGemini.conf.
    data_dir: "/gemini-data/data"
    # [data].store-meta-dir in openGemini.conf.
    meta_dir: "/gemini-data/data/meta"
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"
  - host: 10.0.1.19
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [data].store-ingest-addr in openGemini.conf.
    ingest_port: 8400
    # [data].store-select-addr in openGemini.conf.
    select_port: 8401
    # [gossip].store-bind-port in openGemini.conf.
    gossip_port: 8011
    # [data].store-data-dir & [data].store-wal-dir in openGemini.conf.
    data_dir: "/gemini-data/data"
    # [data].store-meta-dir in openGemini.conf.
    meta_dir: "/gemini-data/data/meta"
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"
```
