package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kainonly/auditstream/v3/common"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type Option struct {
	Key         string   `json:"key"`
	Subs        []string `json:"subs"`
	Stream      string   `json:"stream"`
	Description string   `json:"description"`
}

func (x *App) StreamName(key string) string {
	return fmt.Sprintf("%s_%s", x.V.Namespace, key)
}

func (x *App) SubName(key string) string {
	return fmt.Sprintf("%s.%s", x.V.Namespace, key)
}

func (x *App) Subscribe(option Option) (err error) {
	// 如果已存在，先停止旧的订阅
	if sub := x.Subs.Get(option.Key); sub != nil {
		sub.Stop()
		x.Subs.Delete(option.Key)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var consumer jetstream.Consumer
	if consumer, err = x.Js.Consumer(ctx, x.StreamName(option.Key), "default"); err != nil {
		return
	}

	var cc jetstream.ConsumeContext
	sub := &Subscription{
		option: option,
		app:    x,
		buffer: make([]jetstream.Msg, 0, x.V.BatchSize),
		stopCh: make(chan struct{}),
	}

	if cc, err = consumer.Consume(func(msg jetstream.Msg) {
		sub.Push(msg)
	}); err != nil {
		return
	}

	sub.cc = cc
	go sub.flushLoop()

	x.Subs.Set(option.Key, sub)
	common.Log.Info("subscribe ok",
		zap.String("key", option.Key),
		zap.String("subject", x.SubName(option.Key)),
	)
	return
}

func (x *App) Unsubscribe(key string) (err error) {
	if sub := x.Subs.Get(key); sub != nil {
		sub.Stop()
		x.Subs.Delete(key)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = x.Js.DeleteStream(ctx, x.StreamName(key)); err != nil {
		return
	}

	common.Log.Info("unsubscribe ok",
		zap.String("key", key),
	)
	return
}

// SubscriptionMap 管理所有订阅
type SubscriptionMap struct {
	m sync.Map
}

func NewSubscriptionMap() *SubscriptionMap {
	return &SubscriptionMap{}
}

func (s *SubscriptionMap) Set(key string, sub *Subscription) {
	s.m.Store(key, sub)
}

func (s *SubscriptionMap) Get(key string) *Subscription {
	if v, ok := s.m.Load(key); ok {
		return v.(*Subscription)
	}
	return nil
}

func (s *SubscriptionMap) Delete(key string) {
	s.m.Delete(key)
}

func (s *SubscriptionMap) StopAll() {
	s.m.Range(func(key, value any) bool {
		if sub, ok := value.(*Subscription); ok {
			sub.Stop()
		}
		return true
	})
}
