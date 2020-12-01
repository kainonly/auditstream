package controller

import (
	pb "elastic-transfer/api"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

func (c *controller) Delete(_ context.Context, option *pb.ID) (_ *empty.Empty, err error) {
	if err = c.Transfer.Delete(option.Id); err != nil {
		return
	}
	return &empty.Empty{}, nil
}
