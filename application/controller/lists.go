package controller

import (
	"context"
	pb "elastic-transfer/api"
)

func (c *controller) Lists(_ context.Context, option *pb.IDs) (*pb.DataLists, error) {
	lists := make([]*pb.Data, len(option.Ids))
	for key, val := range option.Ids {
		lists[key] = c.find(val)
	}
	return &pb.DataLists{Data: lists}, nil
}
