package common

import (
	"sync"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

var Log *zap.Logger

// Values is loaded from config/values.yml.
type Values struct {
	Mode        string   `yaml:"mode"`
	Namespace   string   `yaml:"namespace"`
	Description string   `yaml:"description"`
	NatsHosts   []string `yaml:"nats_hosts"`
	NatsToken   string   `yaml:"nats_token"`
	Victoria    string   `yaml:"victoria"`
}

type ConsumerSyncMap struct {
	m sync.Map
}

func NewConsumerSyncMap() *ConsumerSyncMap {
	return &ConsumerSyncMap{}
}

func (c *ConsumerSyncMap) Set(key string, cc jetstream.ConsumeContext) {
	c.m.Store(key, cc)
}

func (c *ConsumerSyncMap) Get(key string) jetstream.ConsumeContext {
	if v, ok := c.m.Load(key); ok {
		return v.(jetstream.ConsumeContext)
	}
	return nil
}

func (c *ConsumerSyncMap) Delete(key string) {
	c.m.Delete(key)
}

func (c *ConsumerSyncMap) StopAll() {
	c.m.Range(func(key, value any) bool {
		if cc, ok := value.(jetstream.ConsumeContext); ok {
			cc.Stop()
		}
		return true
	})
}
