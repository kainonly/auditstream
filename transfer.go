package transfer

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/msgpack/v5"
	"time"
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
		Bucket: namespace,
	}); err != nil {
		return
	}
	return
}

type Option struct {
	Topic       string `msgpack:"topic"`
	Description string `msgpack:"description"`
}

func (x *Transfer) Get(key string) (result map[string]interface{}, err error) {
	result = make(map[string]interface{})
	var b []byte
	if b, err = x.Store.GetBytes(key); err != nil {
		return
	}
	var option Option
	if err = msgpack.Unmarshal(b, &option); err != nil {
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

func (x *Transfer) Set(key string, option Option) (err error) {
	var b []byte
	if b, err = msgpack.Marshal(option); err != nil {
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

func (x *Transfer) Remove(key string) (err error) {
	if err = x.Store.Delete(key); err != nil {
		return
	}
	name := fmt.Sprintf(`%s:logs:%s`, x.Namespace, key)
	return x.Js.DeleteStream(name)
}

type CLSDto struct {
	// CLS 日志主题 ID
	TopicId string `msgpack:"topic_id" validate:"required"`

	// 写入记录
	Record map[string]string `msgpack:"record" validate:"required"`

	// 时间
	Time time.Time `msgpack:"time" validate:"required"`
}

type InfluxDto struct {
	// 度量
	Measurement string `msgpack:"measurement" validate:"required"`

	// 标签
	Tags map[string]string `msgpack:"tags" validate:"required"`

	// 字段
	Fields map[string]interface{} `msgpack:"fields" validate:"required"`

	// 时间
	Time time.Time `msgpack:"time" validate:"required"`
}

type Payload []byte

// NewPayload 创建载荷
func NewPayload(data interface{}) (payload Payload, err error) {
	if err = validator.New().Struct(data); err != nil {
		return
	}
	if payload, err = msgpack.Marshal(data); err != nil {
		return
	}
	return
}

func (x *Transfer) Publish(ctx context.Context, topic string, payload Payload) (err error) {
	subject := fmt.Sprintf(`%s.logs.%s`, x.Namespace, topic)
	if _, err = x.Js.Publish(subject, payload, nats.Context(ctx)); err != nil {
		return
	}
	return
}
