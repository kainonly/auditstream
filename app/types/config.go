package types

import "github.com/elastic/go-elasticsearch/v8"

type Config struct {
	Listen  string               `yaml:"listen"`
	Debug   string               `yaml:"debug"`
	Elastic elasticsearch.Config `yaml:"elastic"`
	Mq      MqOption             `yaml:"mq"`
}
