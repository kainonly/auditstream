package transfer

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"time"
)

type Transfer struct {
	// 命名空间
	Namespace string
	// Nats JetStream
	Js nats.JetStreamContext
	// Nats ObjectStore
	Store nats.ObjectStore
}

// New 新建传输
func New(namespace string, js nats.JetStreamContext) (x *Transfer, err error) {
	x = &Transfer{
		Namespace: namespace,
		Js:        js,
	}
	if x.Store, err = js.CreateObjectStore(&nats.ObjectStoreConfig{
		Bucket: fmt.Sprintf(`%s_logs`, namespace),
	}); err != nil {
		return
	}
	return
}

type Option struct {
	// 度量
	Measurement string `json:"measurement"`
	// 描述
	Description string `json:"description"`
}

// Get 获取日志流传输信息，JetStream 状态
func (x *Transfer) Get(measurement string) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	var b []byte
	if b, err = x.Store.GetBytes(measurement); err != nil {
		return
	}
	var option Option
	if err = sonic.Unmarshal(b, &option); err != nil {
		return
	}
	data["option"] = option
	name := fmt.Sprintf(`%s:logs:%s`, x.Namespace, measurement)
	var info *nats.StreamInfo
	if info, err = x.Js.StreamInfo(name); err != nil {
		return
	}
	data["info"] = *info
	return
}

// Set 设置日志流传输
func (x *Transfer) Set(option Option) (err error) {
	var b []byte
	if b, err = sonic.Marshal(option); err != nil {
		return
	}
	if _, err = x.Store.PutBytes(option.Measurement, b); err != nil {
		return
	}
	name := fmt.Sprintf(`%s:logs:%s`, x.Namespace, option.Measurement)
	subject := fmt.Sprintf(`%s.logs.%s`, x.Namespace, option.Measurement)
	if _, err = x.Js.AddStream(&nats.StreamConfig{
		Name:        name,
		Subjects:    []string{subject},
		Description: option.Description,
		Retention:   nats.WorkQueuePolicy,
	}); err != nil {
		return
	}
	return
}

// Remove 移除日志流传输
func (x *Transfer) Remove(measurement string) (err error) {
	if err = x.Store.Delete(measurement); err != nil {
		return
	}
	name := fmt.Sprintf(`%s:logs:%s`, x.Namespace, measurement)
	return x.Js.DeleteStream(name)
}

// Payload 载荷
type Payload struct {
	// 标签
	Tags map[string]string `json:"tags"`

	// 字段
	Fields map[string]interface{} `json:"fields"`

	// 时间
	Time time.Time `json:"time"`
}

// Publish 发布
func (x *Transfer) Publish(ctx context.Context, measurement string, payload Payload) (err error) {
	var b []byte
	if b, err = sonic.Marshal(payload); err != nil {
		return
	}
	subject := fmt.Sprintf(`%s.logs.%s`, x.Namespace, measurement)
	if _, err = x.Js.Publish(subject, b, nats.Context(ctx)); err != nil {
		return
	}
	return
}
