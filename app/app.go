package app

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kainonly/auditstream/v3/common"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type App struct {
	V  *common.Values
	Js jetstream.JetStream

	cc     jetstream.ConsumeContext
	mu     sync.Mutex
	buffer []jetstream.Msg
	stopCh chan struct{}
}

func New(v *common.Values, js jetstream.JetStream) *App {
	return &App{
		V:      v,
		Js:     js,
		buffer: make([]jetstream.Msg, 0, v.BatchSize),
		stopCh: make(chan struct{}),
	}
}

func (x *App) Run(ctx context.Context) (err error) {
	streamName := fmt.Sprintf("%s_%s", x.V.Namespace, x.V.Stream)

	var consumer jetstream.Consumer
	if consumer, err = x.Js.Consumer(ctx, streamName, "default"); err != nil {
		return
	}

	if x.cc, err = consumer.Consume(func(msg jetstream.Msg) {
		x.push(msg)
	}); err != nil {
		return
	}

	common.Log.Info("consuming stream",
		zap.String("stream", streamName),
	)

	go x.flushLoop()

	<-ctx.Done()
	return
}

func (x *App) push(msg jetstream.Msg) {
	x.mu.Lock()
	x.buffer = append(x.buffer, msg)
	shouldFlush := len(x.buffer) >= x.V.BatchSize
	x.mu.Unlock()

	if shouldFlush {
		x.flush()
	}
}

func (x *App) flushLoop() {
	ticker := time.NewTicker(x.V.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			x.flush()
		case <-x.stopCh:
			x.flush()
			return
		}
	}
}

func (x *App) flush() {
	x.mu.Lock()
	if len(x.buffer) == 0 {
		x.mu.Unlock()
		return
	}
	msgs := x.buffer
	x.buffer = make([]jetstream.Msg, 0, x.V.BatchSize)
	x.mu.Unlock()

	if err := x.writeBatch(msgs); err != nil {
		common.Log.Error("flush fail",
			zap.Int("count", len(msgs)),
			zap.Error(err),
		)
		for _, msg := range msgs {
			_ = msg.Nak()
		}
		return
	}

	for _, msg := range msgs {
		_ = msg.Ack()
	}

	common.Log.Debug("flush ok",
		zap.Int("count", len(msgs)),
	)
}

func (x *App) writeBatch(msgs []jetstream.Msg) error {
	var buf bytes.Buffer
	for _, msg := range msgs {
		buf.Write(msg.Data())
		buf.WriteByte('\n')
	}

	url := x.V.Victoria + "/insert/jsonline?_stream_fields=stream&_msg_field=msg&_time_field=time"
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/stream+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (x *App) Close() {
	if x.cc != nil {
		x.cc.Stop()
	}
	close(x.stopCh)
}
