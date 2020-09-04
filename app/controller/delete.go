package controller

import (
	pb "elastic-transfer/router"
	"golang.org/x/net/context"
)

func (c *controller) Delete(ctx context.Context, param *pb.DeleteParameter) (*pb.Response, error) {
	return nil, nil
}
