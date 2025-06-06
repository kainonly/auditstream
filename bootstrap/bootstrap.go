package bootstrap

import (
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/go-co-op/gocron/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/weplanx/ppcollector/v3/common"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

func SetZap() (log *zap.Logger, err error) {
	if os.Getenv("MODE") != "release" {
		if log, err = zap.NewDevelopment(); err != nil {
			return
		}
	} else {
		if log, err = zap.NewProduction(); err != nil {
			return
		}
	}
	return
}

func LoadStaticValues() (v *common.Values, err error) {
	v = new(common.Values)
	var b []byte
	if b, err = os.ReadFile("./config/values.yml"); err != nil {
		return
	}
	if err = yaml.Unmarshal(b, &v); err != nil {
		return
	}
	return
}

func UseNats(values *common.Values) (nc *nats.Conn, err error) {
	if nc, err = nats.Connect(
		strings.Join(values.Nats.Hosts, ","),
		nats.Token(values.Nats.Token),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
	); err != nil {
		return
	}
	return
}

func UseJetStream(nc *nats.Conn) (jetstream.JetStream, error) {
	return jetstream.New(nc)
}

func UseKeyValue(values *common.Values, js nats.JetStreamContext) (nats.KeyValue, error) {
	return js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket:      values.Namespace,
		Description: values.Description,
	})
}

func UseDatabase(v *common.Values) (conn clickhouse.Conn, err error) {
	return clickhouse.Open(&clickhouse.Options{
		Protocol: 0,
		Addr:     v.Database.Addr,
		Auth: clickhouse.Auth{
			Database: v.Database.Auth.Database,
			Username: v.Database.Auth.Username,
			Password: v.Database.Auth.Password,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		MaxIdleConns: v.Database.MaxIdleConns,
	})
}

func UseSchedule() (gocron.Scheduler, error) {
	return gocron.NewScheduler()
}
