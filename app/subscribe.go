package app

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/kainonly/auditstream/v3/common"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type Option struct {
	Key         string   `json:"key"`
	Subs        []string `json:"subs"`
	Collection  string   `json:"collection"`
	Description string   `json:"description"`
}

func (x *App) StreamName(key string) string {
	return fmt.Sprintf("%s_%s", x.V.Namespace, key)
}

func (x *App) SubName(key string) string {
	return fmt.Sprintf("%s.%s", x.V.Namespace, key)
}

func (x *App) Subscribe(option Option) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var consumer jetstream.Consumer
	if consumer, err = x.Js.Consumer(ctx, x.StreamName(option.Key), "default"); err != nil {
		return
	}

	var job gocron.Job
	if job, err = x.Schedule.NewJob(
		gocron.DurationJob(x.V.Duration),
		gocron.NewTask(func(o Option, c jetstream.Consumer) {
			if errX := x.Task(o, c); errX != nil {
				common.Log.Error("task fail",
					zap.String("key", o.Key),
					zap.Error(errX),
				)
			}
		}, option, consumer),
		gocron.WithTags(option.Key),
	); err != nil {
		return
	}

	x.jobs.Link(option.Key, job)
	common.Log.Info("subscribe ok",
		zap.String("key", option.Key),
		zap.String("subject", x.SubName(option.Key)),
	)
	return
}

func (x *App) Task(option Option, consumer jetstream.Consumer) (err error) {
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var msgBatch jetstream.MessageBatch
	if msgBatch, err = consumer.FetchNoWait(x.V.Batch); err != nil {
		return
	}

	documents := make([]any, 0)
	msgs := make([]jetstream.Msg, 0)
	for msg := range msgBatch.Messages() {
		documents = append(documents, msg.Data())
		msgs = append(msgs, msg)
	}

	if len(documents) == 0 {
		common.Log.Debug("task ok",
			zap.String("key", option.Key),
			zap.Int("documents", len(documents)),
		)
		return
	}

	for _, msg := range msgs {
		msg.Ack()
	}

	common.Log.Info("task ok",
		zap.String("key", option.Key),
		zap.Int("documents", len(documents)),
	)
	return
}

func (x *App) Unsubscribe(key string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = x.Js.DeleteStream(ctx, x.StreamName(key)); err != nil {
		return
	}
	x.jobs.Unlink(key)
	x.Schedule.RemoveByTags(key)
	common.Log.Info("unsubscribe ok",
		zap.String("key", key),
	)
	return
}
