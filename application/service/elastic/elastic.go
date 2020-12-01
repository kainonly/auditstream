package elastic

import (
	"elastic-transfer/application/service/elastic/utils"
	"elastic-transfer/application/service/queue"
	"elastic-transfer/application/service/schema"
	"elastic-transfer/config"
	"elastic-transfer/config/options"
	"errors"
	"github.com/elastic/go-elasticsearch/v8"
	"go.uber.org/fx"
)

type Elastic struct {
	Client *elasticsearch.Client
	Pipes  *utils.PipeMap
	*Dependency
}

type Dependency struct {
	fx.In

	Config *config.Config
	Schema *schema.Schema
	Queue  *queue.Queue
}

var (
	NotExists = errors.New("this identity does not exists")
)

func New(dep *Dependency) (es *Elastic, err error) {
	es = new(Elastic)
	es.Dependency = dep
	if es.Client, err = elasticsearch.NewClient(dep.Config.Elastic); err != nil {
		return
	}
	es.Pipes = utils.NewPipeMap()
	var pipesOptions []options.PipeOption
	if pipesOptions, err = dep.Schema.Lists(); err != nil {
		return
	}
	for _, option := range pipesOptions {
		if err = es.Put(option); err != nil {
			return
		}
	}
	return
}

func (c *Elastic) GetPipe(identity string) (*options.PipeOption, error) {
	if c.Pipes.Empty(identity) {
		return nil, NotExists
	}
	return c.Pipes.Get(identity), nil
}
