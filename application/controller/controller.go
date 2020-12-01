package controller

import (
	pb "elastic-transfer/api"
	"elastic-transfer/application/common"
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

func (c *controller) find(identity string) *pb.Data {
	pipe := c.ES.Pipes.Get(identity)
	return &pb.Data{
		Id:       pipe.Identity,
		Index:    pipe.Index,
		Validate: pipe.Validate,
		Topic:    pipe.Topic,
		Key:      pipe.Key,
	}
}
