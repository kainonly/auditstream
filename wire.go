//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/smallnest/rpcx/server"
	"github.com/weplanx/transfer/app"
	"github.com/weplanx/transfer/bootstrap"
	"github.com/weplanx/transfer/common"
)

func App(value *common.Values) (*server.Server, error) {
	wire.Build(
		bootstrap.Provides,
		app.Provides,
	)
	return &server.Server{}, nil
}
