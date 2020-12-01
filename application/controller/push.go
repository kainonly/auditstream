package controller

import (
	"context"
	pb "elastic-transfer/api"
	"github.com/golang/protobuf/ptypes/empty"
)

func (c *controller) Push(_ context.Context, body *pb.Body) (_ *empty.Empty, err error) {
	if err = c.Transfer.Push(body.Id, body.Content); err != nil {
		return
	}
	return &empty.Empty{}, nil
}
