package manage

import (
	"elastic-transfer/app/mq"
	"elastic-transfer/app/schema"
	"elastic-transfer/app/types"
	"errors"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/panjf2000/ants/v2"
)

type ElasticManager struct {
	client  *elasticsearch.Client
	mq      *mq.MessageQueue
	pipes   map[string]*types.PipeOption
	schema  *schema.Schema
	pool    *ants.Pool
	runtime int64
}

func NewElasticManager(client *elasticsearch.Client, mq *mq.MessageQueue, pool *ants.Pool) (manager *ElasticManager, err error) {
	manager = new(ElasticManager)
	manager.client = client
	manager.mq = mq
	manager.pipes = make(map[string]*types.PipeOption)
	manager.schema = schema.New()
	var pipesOptions []types.PipeOption
	pipesOptions, err = manager.schema.Lists()
	if err != nil {
		return
	}
	for _, option := range pipesOptions {
		err = manager.Put(option)
		if err != nil {
			return
		}
	}
	manager.pool = pool
	manager.runtime = 0
	return
}

func (c *ElasticManager) empty(identity string) error {
	if c.pipes[identity] == nil {
		return errors.New("this identity does not exists")
	}
	return nil
}

func (c *ElasticManager) GetIdentityCollection() []string {
	var keys []string
	for key := range c.pipes {
		keys = append(keys, key)
	}
	return keys
}

func (c *ElasticManager) GetOption(identity string) (option *types.PipeOption, err error) {
	if err = c.empty(identity); err != nil {
		return
	}
	option = c.pipes[identity]
	return
}
