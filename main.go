package main

import (
	"elastic-transfer/application"
	"elastic-transfer/bootstrap"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.NopLogger,
		fx.Provide(
			bootstrap.LoadConfiguration,
			bootstrap.InitializeSchema,
			bootstrap.InitializeElastic,
			bootstrap.InitializeQueue,
			bootstrap.InitializeTransfer,
		),
		fx.Invoke(application.Application),
	).Run()
}
