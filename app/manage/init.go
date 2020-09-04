package manage

import (
	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticManager struct {
	client *elasticsearch.Client
}

func NewElasticManager(config elasticsearch.Config) (manager *ElasticManager, err error) {
	manager = new(ElasticManager)
	manager.client, err = elasticsearch.NewClient(config)
	return
}
