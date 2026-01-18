package app

import (
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kainonly/auditstream/v3/common"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type State struct {
	Nexts []time.Time `json:"nexts"`
	Last  time.Time   `json:"last"`
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

func (x *App) LoadState(key string) (b []byte, err error) {
	job, state := x.jobs.Get(key), new(State)
	if job == nil {
		return sonic.Marshal(state)
	}
	if state.Nexts, err = job.NextRuns(5); err != nil {
		return
	}
	if state.Last, err = job.LastRun(); err != nil {
		return
	}
	return sonic.Marshal(state)
}
