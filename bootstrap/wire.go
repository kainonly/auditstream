//go:build wireinject
// +build wireinject

package bootstrap

import (
	"github.com/google/wire"
	"github.com/weplanx/collector-clickhouse/v3/app"
	"github.com/weplanx/collector-clickhouse/v3/common"
)

func NewApp() (*app.App, error) {
	wire.Build(
		wire.Struct(new(common.Inject), "*"),
		LoadStaticValues,
		UseDatabase,
		UseNats,
		UseJetStream,
		UseKeyValue,
		UseSchedule,
		app.Initialize,
	)
	return &app.App{}, nil
}
