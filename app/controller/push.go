package controller

import (
	"context"
	pb "elastic-transfer/router"
)

func (c *controller) Push(ctx context.Context, param *pb.PushParameter) (*pb.Response, error) {
	err := c.manager.Push(param.Identity, param.Data)
	if err != nil {
		return c.response(err)
	}
	return c.response(nil)
}
