package controller

import (
	pb "elastic-transfer/api"
	"elastic-transfer/application/common"
	"elastic-transfer/config/options"
)

type controller struct {
	pb.UnimplementedAPIServer
	*common.Dependency
}

func New(dep *common.Dependency) *controller {
	c := new(controller)
	c.Dependency = dep
	return c
}

func (c *controller) find(identity string) (_ *pb.Data, err error) {
	var pipe *options.PipeOption
	if pipe, err = c.Transfer.GetPipe(identity); err != nil {
		return
	}
	return &pb.Data{
		Id:       pipe.Identity,
		Index:    pipe.Index,
		Validate: pipe.Validate,
		Topic:    pipe.Topic,
		Key:      pipe.Key,
	}, nil
}
