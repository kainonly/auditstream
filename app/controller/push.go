package controller

import (
	"context"
	pb "elastic-transfer/router"
)

func (c *controller) Push(ctx context.Context, param *pb.PushParameter) (*pb.Response, error) {
	go c.manager.Push(param.Identity, param.Data)
	return c.response(nil)
}
