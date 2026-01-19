package common

import (
	"time"

	"go.uber.org/zap"
)

var Log *zap.Logger

type Values struct {
	Mode          string        `yaml:"mode"`
	Namespace     string        `yaml:"namespace"`
	Stream        string        `yaml:"stream"`
	NatsHosts     []string      `yaml:"nats_hosts"`
	NatsToken     string        `yaml:"nats_token"`
	Victoria      string        `yaml:"victoria"`
	BatchSize     int           `yaml:"batch_size"`
	FlushInterval time.Duration `yaml:"flush_interval"`
}
