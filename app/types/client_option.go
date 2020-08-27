package types

type ClientOption struct {
	Identity   string         `yaml:"identity"`
	Host       string         `yaml:"host"`
	Port       uint32         `yaml:"port"`
	Username   string         `yaml:"username"`
	Password   string         `yaml:"password"`
	Key        string         `yaml:"key"`
	PassPhrase string         `yaml:"pass_phrase"`
	Tunnels    []TunnelOption `yaml:"tunnels"`
}
