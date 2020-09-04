package manage

import (
	"bytes"
	"errors"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type ElasticManager struct {
	client *elasticsearch.Client
}

func NewElasticManager(config elasticsearch.Config) (manager *ElasticManager, err error) {
	manager = new(ElasticManager)
	manager.client, err = elasticsearch.NewClient(config)
	return
}

func (c *ElasticManager) Index(index string, data []byte) (err error) {
	var res *esapi.Response
	res, err = c.client.Index(
		index,
		bytes.NewBuffer(data),
	)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.IsError() {
		return errors.New(res.Status())
	}
	return
}
