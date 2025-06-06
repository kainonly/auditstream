package common

import (
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/go-co-op/gocron/v2"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

var Log *zap.Logger

type Inject struct {
	V        *Values
	Js       jetstream.JetStream
	Kv       jetstream.KeyValue
	Conn     clickhouse.Conn
	Schedule gocron.Scheduler
}
