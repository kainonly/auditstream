package transfer

import (
	"elastic-transfer/application/service/elastic"
	"elastic-transfer/application/service/queue"
	"elastic-transfer/application/service/schema"
	"elastic-transfer/application/service/transfer/utils"
	"elastic-transfer/config"
	"elastic-transfer/config/options"
	"errors"
	"go.uber.org/fx"
)

type Transfer struct {
	Pipes *utils.PipeMap
	*Dependency
}

type Dependency struct {
	fx.In

	Config *config.Config
	Schema *schema.Schema
	ES     *elastic.Elastic
	Queue  *queue.Queue
}

var (
	NotExists = errors.New("this identity does not exists")
)

func New(dep *Dependency) (c *Transfer, err error) {
	c = new(Transfer)
	c.Dependency = dep
	c.Pipes = utils.NewPipeMap()
	var pipesOptions []options.PipeOption
	if pipesOptions, err = dep.Schema.Lists(); err != nil {
		return
	}
	for _, option := range pipesOptions {
		if err = c.Put(option); err != nil {
			return
		}
	}
	return
}

func (c *Transfer) GetPipe(identity string) (*options.PipeOption, error) {
	if c.Pipes.Empty(identity) {
		return nil, NotExists
	}
	return c.Pipes.Get(identity), nil
}
