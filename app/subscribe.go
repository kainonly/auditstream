package app

import (
	"context"
	"fmt"
	"time"

	"github.com/kainonly/auditstream/v3/common"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

func (x *App) StreamName(key string) string {
	return fmt.Sprintf("%s_%s", x.V.Namespace, key)
}

func (x *App) SubName(key string) string {
	return fmt.Sprintf("%s.%s", x.V.Namespace, key)
}

type Option struct {
	Key         string   `json:"key"`
	Subs        []string `json:"subs"`
	Stream      string   `json:"stream"`
	Description string   `json:"description"`
}

func (x *App) Subscribe(option Option) (err error) {
	// 如果已存在，先停止旧的消费者
	if cc := x.Consumers.Get(option.Key); cc != nil {
		cc.Stop()
		x.Consumers.Delete(option.Key)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var consumer jetstream.Consumer
	if consumer, err = x.Js.Consumer(ctx, x.StreamName(option.Key), "default"); err != nil {
		return
	}

	var cc jetstream.ConsumeContext
	if cc, err = consumer.Consume(func(msg jetstream.Msg) {
		x.handleMessage(option, msg)
	}); err != nil {
		return
	}

	x.Consumers.Set(option.Key, cc)
	common.Log.Info("subscribe ok",
		zap.String("key", option.Key),
		zap.String("subject", x.SubName(option.Key)),
	)
	return
}

func (x *App) handleMessage(option Option, msg jetstream.Msg) {
	fmt.Println(option)

	if err := msg.Ack(); err != nil {
		common.Log.Error("ack fail",
			zap.String("key", option.Key),
			zap.Error(err),
		)
		return
	}

	common.Log.Debug("message processed",
		zap.String("key", option.Key),
		zap.ByteString("data", msg.Data()),
	)
}

func (x *App) Unsubscribe(key string) (err error) {
	if cc := x.Consumers.Get(key); cc != nil {
		cc.Stop()
		x.Consumers.Delete(key)
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
