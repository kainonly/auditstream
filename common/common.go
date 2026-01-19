package common

import (
	"time"

	"go.uber.org/zap"
)

var Log *zap.Logger

// Values is loaded from config/values.yml.
type Values struct {
	Mode          string        `yaml:"mode"`
	Namespace     string        `yaml:"namespace"`
	Description   string        `yaml:"description"`
	NatsHosts     []string      `yaml:"nats_hosts"`
	NatsToken     string        `yaml:"nats_token"`
	Victoria      string        `yaml:"victoria"`
	BatchSize     int           `yaml:"batch_size"`
	FlushInterval time.Duration `yaml:"flush_interval"`
}
