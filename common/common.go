package common

type Values struct {
	Address   string   `yaml:"address"`
	TLS       TLS      `yaml:"tls"`
	Namespace string   `yaml:"namespace"`
	Database  Database `yaml:"database"`
	Nats      Nats     `yaml:"nats"`
}

type TLS struct {
	Ca   string `yaml:"ca"`
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type Database struct {
	Uri        string `yaml:"uri"`
	Name       string `yaml:"name"`
	Collection string `yaml:"collection"`
}

type Nats struct {
	Hosts []string `yaml:"hosts"`
	Nkey  string   `yaml:"nkey"`
}
