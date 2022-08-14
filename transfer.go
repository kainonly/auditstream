package transfer

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
)

type Transfer struct {
	Namespace string
	Js        nats.JetStreamContext
	Store     nats.ObjectStore
}

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
	// 主题
	Topic string `json:"topic"`
	// 描述
	Description string `json:"description"`
}

// Get 获取传输器信息
func (x *Transfer) Get(key string) (result map[string]interface{}, err error) {
	result = make(map[string]interface{})
	var b []byte
	if b, err = x.Store.GetBytes(key); err != nil {
		return
	}
	var option Option
	if err = sonic.Unmarshal(b, &option); err != nil {
		return
	}
	result["option"] = option
	name := fmt.Sprintf(`%s:logs:%s`, x.Namespace, key)
	var info *nats.StreamInfo
	if info, err = x.Js.StreamInfo(name); err != nil {
		return
	}
	result["info"] = *info
	return
}

// Set 设置传输器
func (x *Transfer) Set(key string, option Option) (err error) {
	var b []byte
	if b, err = sonic.Marshal(option); err != nil {
		return
	}
	if _, err = x.Store.PutBytes(key, b); err != nil {
		return
	}
	name := fmt.Sprintf(`%s:logs:%s`, x.Namespace, key)
	subject := fmt.Sprintf(`%s.logs.%s`, x.Namespace, option.Topic)
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

// Remove 移除配置
func (x *Transfer) Remove(key string) (err error) {
	if err = x.Store.Delete(key); err != nil {
		return
	}
	name := fmt.Sprintf(`%s:logs:%s`, x.Namespace, key)
	return x.Js.DeleteStream(name)
}

// Publish 发布
func (x *Transfer) Publish(ctx context.Context, topic string, payload map[string]interface{}) (err error) {
	var b []byte
	if b, err = sonic.Marshal(payload); err != nil {
		return
	}
	subject := fmt.Sprintf(`%s.logs.%s`, x.Namespace, topic)
	if _, err = x.Js.Publish(subject, b, nats.Context(ctx)); err != nil {
		return
	}
	return
}
