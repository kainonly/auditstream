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

// App 是审计日志消费服务的核心结构
// 负责从 JetStream 消费消息，缓冲后批量写入 VictoriaLogs
type App struct {
	V  *common.Values
	Js jetstream.JetStream

	cc     jetstream.ConsumeContext // JetStream 消费上下文，用于停止消费
	mu     sync.Mutex               // 保护 buffer 的并发访问
	buffer []jetstream.Msg          // 消息缓冲区
	stopCh chan struct{}            // 停止信号通道
}

// New 创建 App 实例
// 初始化 buffer 容量为 BatchSize，避免频繁扩容
func New(v *common.Values, js jetstream.JetStream) *App {
	return &App{
		V:      v,
		Js:     js,
		buffer: make([]jetstream.Msg, 0, v.BatchSize),
		stopCh: make(chan struct{}),
	}
}

// Run 启动消费循环
// 1. 获取 JetStream consumer
// 2. 启动 push-based 消费
// 3. 启动定时 flush 协程
// 4. 阻塞直到 ctx 取消
func (x *App) Run(ctx context.Context) (err error) {
	streamName := fmt.Sprintf("%s_%s", x.V.Namespace, x.V.Stream)

	var consumer jetstream.Consumer
	if consumer, err = x.Js.Consumer(ctx, streamName, "default"); err != nil {
		return
	}

	// Consume 是 push-based 模式，消息到达时自动调用回调
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

// push 将消息加入缓冲区
// 如果缓冲区达到 BatchSize，立即触发 flush
func (x *App) push(msg jetstream.Msg) {
	x.mu.Lock()
	x.buffer = append(x.buffer, msg)
	shouldFlush := len(x.buffer) >= x.V.BatchSize
	x.mu.Unlock()

	// 达到批量大小，立即 flush（不等待定时器）
	if shouldFlush {
		x.flush()
	}
}

// flushLoop 定时 flush 循环
// 每隔 FlushInterval 触发一次 flush
// 收到停止信号时执行最后一次 flush 后退出
func (x *App) flushLoop() {
	ticker := time.NewTicker(x.V.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			x.flush()
		case <-x.stopCh:
			x.flush() // 停止前最后 flush 一次，确保不丢消息
			return
		}
	}
}

// flush 将缓冲区消息批量写入 VictoriaLogs
// 使用 swap buffer 技术减少锁持有时间
// 写入成功: ACK 所有消息
// 写入失败: NAK 所有消息，让它们重新入队
func (x *App) flush() {
	// swap buffer: 快速交换出当前缓冲区，减少锁持有时间
	x.mu.Lock()
	if len(x.buffer) == 0 {
		x.mu.Unlock()
		return
	}
	msgs := x.buffer
	x.buffer = make([]jetstream.Msg, 0, x.V.BatchSize)
	x.mu.Unlock()

	// 写入 VictoriaLogs
	if err := x.writeBatch(msgs); err != nil {
		common.Log.Error("flush fail",
			zap.Int("count", len(msgs)),
			zap.Error(err),
		)
		// 写入失败，NAK 让消息重新入队
		for _, msg := range msgs {
			_ = msg.Nak()
		}
		return
	}

	// 写入成功，ACK 确认消息
	for _, msg := range msgs {
		_ = msg.Ack()
	}

	common.Log.Debug("flush ok",
		zap.Int("count", len(msgs)),
	)
}

// writeBatch 批量写入 VictoriaLogs
// 使用 JSONL 格式（每行一个 JSON）
func (x *App) writeBatch(msgs []jetstream.Msg) error {
	// 构建 JSONL 格式: 每条消息一行
	var buf bytes.Buffer
	for _, msg := range msgs {
		buf.Write(msg.Data())
		buf.WriteByte('\n')
	}

	// POST 到 VictoriaLogs jsonline 接口
	url := x.V.Victoria + x.V.VictoriaPath
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

// Close 优雅关闭
// 1. 停止 JetStream 消费
// 2. 发送停止信号，触发最后一次 flush
func (x *App) Close() {
	if x.cc != nil {
		x.cc.Stop()
	}
	close(x.stopCh)
}
