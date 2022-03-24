package client

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/weplanx/transfer/api"
	"google.golang.org/grpc"
	"time"
)

type Transfer struct {
	client api.APIClient
	conn   *grpc.ClientConn
}

func New(addr string, opts ...grpc.DialOption) (x *Transfer, err error) {
	x = new(Transfer)
	if x.conn, err = grpc.Dial(addr, opts...); err != nil {
		return
	}
	x.client = api.NewAPIClient(x.conn)
	return
}

func (x *Transfer) Close() error {
	return x.conn.Close()
}

// Get 获取设置
func (x *Transfer) Get(ctx context.Context) (result []*api.Transfer, err error) {
	var rep *api.GetReply
	if rep, err = x.client.Get(ctx, &empty.Empty{}); err != nil {
		return
	}
	return rep.GetData(), nil
}

// Create 创建传输
func (x *Transfer) Create(ctx context.Context, key string, topic string, description string) (err error) {
	if _, err = x.client.Create(ctx,
		&api.CreateRequest{
			Key:         key,
			Topic:       topic,
			Description: description,
		}); err != nil {
		return
	}
	return
}

// Update 更新传输
func (x *Transfer) Update(ctx context.Context, key string, description string) (err error) {
	if _, err = x.client.Update(ctx,
		&api.UpdateRequest{
			Key:         key,
			Description: description,
		}); err != nil {
		return
	}
	return
}

// Delete 删除传输
func (x *Transfer) Delete(ctx context.Context, key string) (err error) {
	if _, err = x.client.Delete(ctx, &api.DeleteRequest{Key: key}); err != nil {
		return
	}
	return
}

// Info 获取日志主题详情
func (x *Transfer) Info(ctx context.Context, key string) (*api.InfoReply, error) {
	return x.client.Info(ctx, &api.InfoRequest{Key: key})
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
func NewPayload[T CLSDto | InfluxDto](data T) (payload Payload, err error) {
	if err = validator.New().Struct(&data); err != nil {
		return
	}
	if payload, err = msgpack.Marshal(data); err != nil {
		return
	}
	return
}

// Publish 投递日志
func (x *Transfer) Publish(ctx context.Context, topic string, payload Payload) (err error) {
	if _, err = x.client.Publish(ctx,
		&api.PublishRequest{
			Topic:   topic,
			Payload: payload,
		}); err != nil {
		return
	}
	return
}
