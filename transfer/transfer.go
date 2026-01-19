package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Transfer 是审计日志发送客户端
// 用于将审计事件推送到 NATS JetStream
type Transfer struct {
	namespace string
	js        jetstream.JetStream
}

// New 创建 Transfer 实例
// namespace: 命名空间，用于构建 subject（格式：{namespace}.{stream}）
func New(nc *nats.Conn, namespace string, opts ...jetstream.JetStreamOpt) (x *Transfer, err error) {
	x = &Transfer{namespace: namespace}
	if x.js, err = jetstream.New(nc, opts...); err != nil {
		return nil, err
	}
	return
}

// Subject 返回完整的 subject 名称
// 格式：{namespace}.{stream}
func (x *Transfer) Subject(stream string) string {
	return fmt.Sprintf("%s.%s", x.namespace, stream)
}

// Publish 同步发送消息到指定 stream
// stream: stream 名称
// data: 消息内容，会被 JSON 序列化
func (x *Transfer) Publish(ctx context.Context, stream string, data any) error {
	content, err := sonic.Marshal(data)
	if err != nil {
		return err
	}
	_, err = x.js.Publish(ctx, x.Subject(stream), content)
	return err
}

// PublishAsync 异步发送消息到指定 stream
// 返回 PubAckFuture，可用于后续确认发送结果
func (x *Transfer) PublishAsync(stream string, data any) (jetstream.PubAckFuture, error) {
	content, err := sonic.Marshal(data)
	if err != nil {
		return nil, err
	}
	return x.js.PublishAsync(x.Subject(stream), content)
}

// PublishRaw 发送原始字节数据（已序列化的 JSON）
func (x *Transfer) PublishRaw(ctx context.Context, stream string, data []byte) error {
	_, err := x.js.Publish(ctx, x.Subject(stream), data)
	return err
}

// PublishRawAsync 异步发送原始字节数据
func (x *Transfer) PublishRawAsync(stream string, data []byte) (jetstream.PubAckFuture, error) {
	return x.js.PublishAsync(x.Subject(stream), data)
}

// AuditEvent 审计事件结构
// 这是推荐的审计日志格式，与 VictoriaLogs 的字段映射对应
type AuditEvent struct {
	Time   time.Time `json:"time"`   // 事件时间（_time_field）
	Stream string    `json:"stream"` // 日志流标识（_stream_fields）
	Msg    string    `json:"msg"`    // 消息内容（_msg_field）

	// 核心查询字段（必填）
	Platform string `json:"platform"`  // 平台标识
	Resource string `json:"resource"`  // 资源类型
	Action   string `json:"action"`    // 操作类型
	ObjectId any    `json:"object_id"` // 对象 ID（主键、外键、数组索引等）
	UserId   int    `json:"user_id"`   // 用户 ID
	Path     string `json:"path"`      // 请求路径
	IP       string `json:"ip"`        // 客户端 IP

	// 扩展字段（用于显示，低频查询）
	Extra map[string]any `json:"extra,omitempty"`
}

// NewAuditEvent 创建审计事件
func NewAuditEvent(stream, msg string) *AuditEvent {
	return &AuditEvent{
		Time:   time.Now(),
		Stream: stream,
		Msg:    msg,
	}
}

// WithMeta 设置元数据
func (e *AuditEvent) WithMeta(platform, resource, action string, objectId any, userId int) *AuditEvent {
	e.Platform = platform
	e.Resource = resource
	e.Action = action
	e.ObjectId = objectId
	e.UserId = userId
	return e
}

// WithRequest 设置请求信息（body 存入 extra）
func (e *AuditEvent) WithRequest(path string, body map[string]any) *AuditEvent {
	e.Path = path
	if body != nil {
		e.ensureExtra()
		e.Extra["body"] = body
	}
	return e
}

// WithResponse 设置响应信息（存入 extra）
func (e *AuditEvent) WithResponse(code int, response any) *AuditEvent {
	e.ensureExtra()
	e.Extra["code"] = code
	if response != nil {
		e.Extra["response"] = response
	}
	return e
}

// WithClient 设置客户端信息（agent 存入 extra）
func (e *AuditEvent) WithClient(ip, agent string) *AuditEvent {
	e.IP = ip
	if agent != "" {
		e.ensureExtra()
		e.Extra["agent"] = agent
	}
	return e
}

// WithExtra 设置额外数据
func (e *AuditEvent) WithExtra(key string, value any) *AuditEvent {
	e.ensureExtra()
	e.Extra[key] = value
	return e
}

func (e *AuditEvent) ensureExtra() {
	if e.Extra == nil {
		e.Extra = make(map[string]any)
	}
}
