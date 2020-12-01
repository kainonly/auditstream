package controller

import (
	"context"
	pb "elastic-transfer/api"
)

func (c *controller) Get(_ context.Context, option *pb.ID) (*pb.Data, error) {
	return c.find(option.Id)
}
