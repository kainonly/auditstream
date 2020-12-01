package queue

import (
	"elastic-transfer/application/service/queue/drive"
)

type Queue struct {
	Drive interface{}
	drive.API
}

type Option struct {
	Drive  string                 `yaml:"drive"`
	Option map[string]interface{} `yaml:"option"`
}

func (c *Queue) Publish(topic string, key string, data []byte) (err error) {
	return c.Drive.(drive.API).Publish(topic, key, data)
}
