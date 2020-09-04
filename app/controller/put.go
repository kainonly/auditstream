package controller

import (
	"context"
	"elastic-transfer/app/types"
	pb "elastic-transfer/router"
)

func (c *controller) Put(ctx context.Context, param *pb.Information) (*pb.Response, error) {
	err := c.manager.Put(types.PipeOption{
		Identity: param.Identity,
		Index:    param.Index,
		Validate: param.Validate,
		Topic:    param.Topic,
		Key:      param.Key,
	})
	if err != nil {
		return c.response(err)
	}
	return c.response(nil)
}
