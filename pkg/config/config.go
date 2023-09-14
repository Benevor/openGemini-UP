package config

type CommonConfig struct {
	MetaHosts  []string //IPs
	StoreHosts []string
	SqlHosts   []string
}

type SSHConfig struct {
	Port       int
	User       string
	Typ        string
	Password   string
	KeyPath    string
	UpDataPath string
}

type Config struct {
	CommonConfig *CommonConfig
	SSHConfig    map[string]SSHConfig
}

type Configurator interface {
	Run() error
	RunWithoutGen() error
	GetConfig() *Config
}

type GeminiConfigurator struct {
	yamlPath string
	tomlPath string
	genPath  string
	version  string
	conf     *Config
}

func NewGeminiConfigurator(yPath, tPath, gPath, v string) Configurator {
	return &GeminiConfigurator{
		yamlPath: yPath,
		tomlPath: tPath,
		genPath:  gPath,
		version:  v,
		conf: &Config{
			CommonConfig: &CommonConfig{},
			SSHConfig:    make(map[string]SSHConfig),
		},
	}
}

func (c *GeminiConfigurator) Run() error {
	var err error
	var t Toml
	var y Yaml
	if y, err = ReadFromYaml(c.yamlPath); err != nil {
		return err
	}
	if t, err = ReadFromToml(c.tomlPath); err != nil {
		return err
	}
	GenConfs(y, t, c.genPath)
	c.buildFromYaml(y)
	return err
}

func (c *GeminiConfigurator) RunWithoutGen() error {
	var err error
	var y Yaml
	if y, err = ReadFromYaml(c.yamlPath); err != nil {
		return err
	}
	c.buildFromYaml(y)
	return err
}

func (c *GeminiConfigurator) GetConfig() *Config {
	return c.conf
}

func (c *GeminiConfigurator) buildFromYaml(y Yaml) {
	for _, meta := range y.TsMeta {
		ssh, ok := c.conf.SSHConfig[meta.Host]
		if !ok {
			ssh = SSHConfig{}
		}
		if meta.SSHPort != 0 {
			ssh.Port = meta.SSHPort
		}
		if meta.DeployDir != "" {
			ssh.UpDataPath = meta.DeployDir
		}
		c.conf.SSHConfig[meta.Host] = ssh

		c.conf.CommonConfig.MetaHosts = append(c.conf.CommonConfig.MetaHosts, meta.Host)
	}
	for _, sql := range y.TsSql {
		ssh, ok := c.conf.SSHConfig[sql.Host]
		if !ok {
			ssh = SSHConfig{}
		}
		if sql.SSHPort != 0 {
			ssh.Port = sql.SSHPort
		}
		if sql.DeployDir != "" {
			ssh.UpDataPath = sql.DeployDir
		}
		c.conf.SSHConfig[sql.Host] = ssh

		c.conf.CommonConfig.SqlHosts = append(c.conf.CommonConfig.SqlHosts, sql.Host)
	}
	for _, store := range y.TsStore {
		ssh, ok := c.conf.SSHConfig[store.Host]
		if !ok {
			ssh = SSHConfig{}
		}
		if store.SSHPort != 0 {
			ssh.Port = store.SSHPort
		}
		if store.DeployDir != "" {
			ssh.UpDataPath = store.DeployDir
		}
		c.conf.SSHConfig[store.Host] = ssh

		c.conf.CommonConfig.StoreHosts = append(c.conf.CommonConfig.StoreHosts, store.Host)
	}
}
