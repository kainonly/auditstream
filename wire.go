//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/weplanx/transfer/api"
	"github.com/weplanx/transfer/bootstrap"
	"github.com/weplanx/transfer/common"
	"google.golang.org/grpc"
)

func App(value *common.Values) (*grpc.Server, error) {
	wire.Build(
		wire.Struct(new(common.Inject), "*"),
		bootstrap.Provides,
		api.Provides,
	)
	return &grpc.Server{}, nil
}
