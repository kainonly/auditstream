package app

import (
	"bytes"
	"net/http"
	"sync"
	"time"

	"github.com/kainonly/auditstream/v3/common"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type Subscription struct {
	option Option
	cc     jetstream.ConsumeContext
	app    *App

	mu     sync.Mutex
	buffer []jetstream.Msg
	stopCh chan struct{}
}

func NewSubscription(app *App, option Option, cc jetstream.ConsumeContext) *Subscription {
	s := &Subscription{
		option: option,
		cc:     cc,
		app:    app,
		buffer: make([]jetstream.Msg, 0, app.V.BatchSize),
		stopCh: make(chan struct{}),
	}
	go s.flushLoop()
	return s
}

func (s *Subscription) Push(msg jetstream.Msg) {
	s.mu.Lock()
	s.buffer = append(s.buffer, msg)
	shouldFlush := len(s.buffer) >= s.app.V.BatchSize
	s.mu.Unlock()

	if shouldFlush {
		s.Flush()
	}
}

func (s *Subscription) flushLoop() {
	ticker := time.NewTicker(s.app.V.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.Flush()
		case <-s.stopCh:
			s.Flush() // 停止前最后 flush 一次
			return
		}
	}
}

func (s *Subscription) Flush() {
	s.mu.Lock()
	if len(s.buffer) == 0 {
		s.mu.Unlock()
		return
	}
	msgs := s.buffer
	s.buffer = make([]jetstream.Msg, 0, s.app.V.BatchSize)
	s.mu.Unlock()

	if err := s.writeBatch(msgs); err != nil {
		common.Log.Error("flush fail",
			zap.String("key", s.option.Key),
			zap.Int("count", len(msgs)),
			zap.Error(err),
		)
		// 写入失败时 NAK 消息，让它们重新入队
		for _, msg := range msgs {
			_ = msg.Nak()
		}
		return
	}

	// 写入成功后 ACK
	for _, msg := range msgs {
		if err := msg.Ack(); err != nil {
			common.Log.Error("ack fail",
				zap.String("key", s.option.Key),
				zap.Error(err),
			)
		}
	}

	common.Log.Debug("flush ok",
		zap.String("key", s.option.Key),
		zap.Int("count", len(msgs)),
	)
}

func (s *Subscription) writeBatch(msgs []jetstream.Msg) error {
	// 构建 JSONL 格式的批量数据
	var buf bytes.Buffer
	for _, msg := range msgs {
		buf.Write(msg.Data())
		buf.WriteByte('\n')
	}

	// 写入 VictoriaLogs
	url := s.app.V.Victoria + "/insert/jsonline?_stream_fields=stream&_msg_field=msg&_time_field=time"
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

func (s *Subscription) Stop() {
	s.cc.Stop()
	close(s.stopCh)
}

func (s *Subscription) IsActive() bool {
	return s.cc != nil
}
