package common

import (
	"time"

	"go.uber.org/zap"
)

var Log *zap.Logger

// Values is loaded from config/values.yml.
type Values struct {
	Mode         string        `yaml:"mode"`
	Namespace    string        `yaml:"namespace"`
	Description  string        `yaml:"description"`
	Duration     time.Duration `yaml:"duration"`
	Batch        int           `yaml:"batch"`
	NatsHosts    []string      `yaml:"nats_hosts"`
	NatsToken    string        `yaml:"nats_token"`
	Victorialogs string        `yaml:"victorialogs"`
}
