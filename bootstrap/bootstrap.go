package bootstrap

import (
	"context"
	"os"
	"strings"

	"github.com/go-co-op/gocron/v2"
	"github.com/kainonly/auditstream/v3/common"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
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

// LoadStaticValues loads configuration from ./config/values.yml.
func LoadStaticValues(path string) (v *common.Values, err error) {
	v = new(common.Values)
	var b []byte
	if b, err = os.ReadFile(path); err != nil {
		return
	}
	if err = yaml.Unmarshal(b, &v); err != nil {
		return
	}
	return
}

// UseNats creates a NATS connection with infinite reconnect attempts.
func UseNats(values *common.Values) (nc *nats.Conn, err error) {
	if nc, err = nats.Connect(
		strings.Join(values.NatsHosts, ","),
		nats.Token(values.NatsToken),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
	); err != nil {
		return
	}
	return
}

// UseJetStream creates a JetStream context.
func UseJetStream(nc *nats.Conn) (jetstream.JetStream, error) {
	return jetstream.New(nc)
}

// UseKeyValue creates or updates the namespace KV bucket used for stream configuration.
func UseKeyValue(values *common.Values, js jetstream.JetStream) (jetstream.KeyValue, error) {
	return js.CreateOrUpdateKeyValue(context.TODO(), jetstream.KeyValueConfig{
		Bucket:      values.Namespace,
		Description: values.Description,
		History:     3,
		Compression: true,
	})
}

// UseSchedule creates a scheduler instance.
func UseSchedule() (gocron.Scheduler, error) {
	return gocron.NewScheduler()
}
