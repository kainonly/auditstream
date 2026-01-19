package app

import (
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/kainonly/auditstream/v3/common"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type State struct {
	Active bool `json:"active"`
}

func (x *App) LoadState(key string) (b []byte, err error) {
	sub := x.Subs.Get(key)
	state := &State{
		Active: sub != nil && sub.IsActive(),
	}
	return sonic.Marshal(state)
}

func (x *App) States() (err error) {
	if _, err = x.Nc.Subscribe(fmt.Sprintf("%s.states", x.V.Namespace), func(m *nats.Msg) {
		key := string(m.Data)
		b, errX := x.LoadState(key)
		if errX != nil {
			common.Log.Error("load state fail",
				zap.String("key", key),
				zap.Error(errX),
			)
			return
		}

		if errX = x.Nc.Publish(m.Reply, b); errX != nil {
			common.Log.Error("publish state fail",
				zap.String("key", key),
				zap.Error(errX),
			)
		}
	}); err != nil {
		return
	}
	return
}
