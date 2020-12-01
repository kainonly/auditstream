package controller

import (
	"context"
	pb "elastic-transfer/api"
	"elastic-transfer/config/options"
	"github.com/golang/protobuf/ptypes/empty"
)

func (c *controller) Put(_ context.Context, Data *pb.Data) (_ *empty.Empty, err error) {
	if err = c.Transfer.Put(options.PipeOption{
		Identity: Data.Id,
		Index:    Data.Index,
		Validate: Data.Validate,
		Topic:    Data.Topic,
		Key:      Data.Key,
	}); err != nil {
		return
	}
	return &empty.Empty{}, nil
}
