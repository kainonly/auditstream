package app

import (
	"context"
	"errors"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kainonly/auditstream/v3/common"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type App struct {
	V  *common.Values
	Nc *nats.Conn
	Js jetstream.JetStream
	Kv jetstream.KeyValue

	Subs *SubscriptionMap
}

func New(v *common.Values, nc *nats.Conn, js jetstream.JetStream, kv jetstream.KeyValue) *App {
	return &App{
		V:    v,
		Nc:   nc,
		Js:   js,
		Kv:   kv,
		Subs: NewSubscriptionMap(),
	}
}

func (x *App) Run(ctx context.Context) (err error) {
	var keys []string
	if keys, err = x.Kv.Keys(ctx); err != nil {
		if errors.Is(err, jetstream.ErrNoKeysFound) {
			keys = make([]string, 0)
		} else {
			return
		}
	}

	for _, key := range keys {
		var entry jetstream.KeyValueEntry
		if entry, err = x.Kv.Get(ctx, key); err != nil {
			return
		}
		var option Option
		if err = sonic.Unmarshal(entry.Value(), &option); err != nil {
			common.Log.Error("decoding fail",
				zap.String("key", key),
				zap.Error(err),
			)
			return
		}
		if err = x.Subscribe(option); err != nil {
			common.Log.Error("subscribe fail",
				zap.String("key", key),
				zap.Error(err),
			)
		}
	}

	common.Log.Info("service initialized successfully.")

	var watch jetstream.KeyWatcher
	if watch, err = x.Kv.WatchAll(ctx); err != nil {
		return
	}

	common.Log.Info("automatically observing configuration changes.")
	cur := time.Now()
	for entry := range watch.Updates() {
		if entry == nil || entry.Created().Unix() < cur.Unix() {
			continue
		}
		key := entry.Key()
		switch entry.Operation() {
		case jetstream.KeyValuePut:
			var option Option
			if err = sonic.Unmarshal(entry.Value(), &option); err != nil {
				common.Log.Error("decoding fail",
					zap.ByteString("data", entry.Value()),
					zap.Error(err),
				)
				return
			}
			if err = x.Subscribe(option); err != nil {
				common.Log.Error("subscribe fail",
					zap.String("key", key),
					zap.Error(err),
				)
			}
		case jetstream.KeyValueDelete, jetstream.KeyValuePurge:
			if err = x.Unsubscribe(key); err != nil {
				common.Log.Error("unsubscribe fail",
					zap.String("key", key),
					zap.Error(err),
				)
			}
		}
	}

	return
}

func (x *App) Close() {
	x.Subs.StopAll()
}
