package transfer

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Transfer struct {
	Namespace string
	Db        *mongo.Database
	Js        nats.JetStreamContext
	KeyValue  nats.KeyValue
}

// New 新建传输
func New(options ...Option) (x *Transfer, err error) {
	x = new(Transfer)
	for _, v := range options {
		v(x)
	}
	if x.KeyValue, err = x.Js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket: fmt.Sprintf(`%s_logs`, x.Namespace),
	}); err != nil {
		return
	}
	return
}

type Option func(x *Transfer)

func SetNamespace(v string) Option {
	return func(x *Transfer) {
		x.Namespace = v
	}
}

func SetDatabase(v *mongo.Database) Option {
	return func(x *Transfer) {
		x.Db = v
	}
}

func SetJetStream(v nats.JetStreamContext) Option {
	return func(x *Transfer) {
		x.Js = v
	}
}

type LogOption struct {
	// 日志标识
	Key string `json:"key"`
	// 描述
	Description string `json:"description"`
	// 有效期
	TTL int64 `json:"ttl"`
}

// Get 获取日志流传输信息，JetStream 状态
func (x *Transfer) Get(key string) (result map[string]interface{}, err error) {
	result = make(map[string]interface{})
	var entry nats.KeyValueEntry
	if entry, err = x.KeyValue.Get(key); err != nil {
		return
	}
	var option LogOption
	if err = sonic.Unmarshal(entry.Value(), &option); err != nil {
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

// Set 设置日志流传输
func (x *Transfer) Set(ctx context.Context, option LogOption) (err error) {
	var b []byte
	if b, err = sonic.Marshal(option); err != nil {
		return
	}
	if _, err = x.KeyValue.Put(option.Key, b); err != nil {
		return
	}

	coll := fmt.Sprintf(`%s_logs`, option.Key)
	var exists []*mongo.CollectionSpecification
	if exists, err = x.Db.ListCollectionSpecifications(ctx, bson.M{"name": coll}); err != nil {
		return
	}

	if len(exists) != 0 {
		if exists[0].Type != "timeseries" {
			return fmt.Errorf(`the [%s] collection already exists, but it must be a timeseries`, coll)
		}
	} else {
		if err = x.Db.CreateCollection(ctx, coll, options.CreateCollection().
			SetTimeSeriesOptions(
				options.TimeSeries().
					SetMetaField("metadata").
					SetTimeField("timestamp"),
			).SetExpireAfterSeconds(option.TTL)); err != nil {
			return
		}
	}

	name := fmt.Sprintf(`%s:logs:%s`, x.Namespace, option.Key)
	subject := fmt.Sprintf(`%s.logs.%s`, x.Namespace, option.Key)
	if _, err = x.Js.AddStream(&nats.StreamConfig{
		Name:        name,
		Subjects:    []string{subject},
		Description: option.Description,
		Retention:   nats.WorkQueuePolicy,
	}, nats.Context(ctx)); err != nil {
		if err = x.Db.Collection(coll).Drop(ctx); err != nil {
			return
		}
		return
	}

	return
}

// Remove 移除日志流传输
func (x *Transfer) Remove(key string) (err error) {
	if err = x.KeyValue.Delete(key); err != nil {
		return
	}
	name := fmt.Sprintf(`%s:logs:%s`, x.Namespace, key)
	return x.Js.DeleteStream(name)
}

type Payload struct {
	// 元数据
	Metadata map[string]interface{} `bson:"metadata" json:"metadata"`
	// 日志
	Data map[string]interface{} `bson:"data" json:"data"`
	// 时间
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

// Publish 发布
func (x *Transfer) Publish(ctx context.Context, key string, payload Payload) (err error) {
	var b []byte
	if b, err = sonic.Marshal(payload); err != nil {
		return
	}
	subject := fmt.Sprintf(`%s.logs.%s`, x.Namespace, key)
	if _, err = x.Js.Publish(subject, b, nats.Context(ctx)); err != nil {
		return
	}
	return
}
