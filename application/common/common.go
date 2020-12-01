package common

import (
	"elastic-transfer/application/service/elastic"
	"elastic-transfer/application/service/queue"
	"elastic-transfer/application/service/schema"
	"elastic-transfer/config"
	"go.uber.org/fx"
)

type Dependency struct {
	fx.In

	Config *config.Config
	Schema *schema.Schema
	Queue  *queue.Queue
	ES     *elastic.Elastic
}
