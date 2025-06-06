package common

import "time"

type Values struct {
	Mode        string        `yaml:"mode"`
	Namespace   string        `yaml:"namespace"`
	Description string        `yaml:"description"`
	Duration    time.Duration `yaml:"duration"`
	Nats        Nats          `yaml:"nats"`
	Database    Database      `yaml:"database"`
}

type Nats struct {
	Hosts []string `yaml:"hosts"`
	Token string   `yaml:"token"`
}

type Database struct {
	Addr         []string      `yaml:"addr"`
	Auth         *DatabaseAuth `yaml:"auth"`
	Name         string        `yaml:"name"`
	MaxIdleConns int           `yaml:"max_idle_conns"`
}

type DatabaseAuth struct {
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
