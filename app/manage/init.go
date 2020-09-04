package manage

import (
	"elastic-transfer/app/schema"
	"elastic-transfer/app/types"
	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticManager struct {
	client *elasticsearch.Client
	pipes  map[string]*types.PipeOption
	schema *schema.Schema
}

func NewElasticManager(config elasticsearch.Config) (manager *ElasticManager, err error) {
	manager = new(ElasticManager)
	manager.client, err = elasticsearch.NewClient(config)
	manager.schema = schema.New()
	return
}
