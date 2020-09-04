package manage

import (
	"elastic-transfer/app/mq"
	"elastic-transfer/app/schema"
	"elastic-transfer/app/types"
	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticManager struct {
	client *elasticsearch.Client
	mq     *mq.MessageQueue
	pipes  map[string]*types.PipeOption
	schema *schema.Schema
}

func NewElasticManager(config elasticsearch.Config, mqOption types.MqOption) (manager *ElasticManager, err error) {
	manager = new(ElasticManager)
	manager.client, err = elasticsearch.NewClient(config)
	manager.mq, err = mq.NewMessageQueue(mqOption)
	if err != nil {
		return
	}
	manager.pipes = make(map[string]*types.PipeOption)
	manager.schema = schema.New()
	return
}
