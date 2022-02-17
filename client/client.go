package client

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/weplanx/transfer/api"
	"google.golang.org/grpc"
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

// GetLoggers 获取日志主题设置
func (x *Transfer) GetLoggers(ctx context.Context) (result []*api.Logger, err error) {
	var rep *api.GetLoggersReply
	if rep, err = x.client.GetLoggers(ctx, &empty.Empty{}); err != nil {
		return
	}
	return rep.GetData(), nil
}

// CreateLogger 创建日志主题
// @topic 主题
// @description 描述
func (x *Transfer) CreateLogger(ctx context.Context, topic string, description string) (err error) {
	if _, err = x.client.CreateLogger(ctx,
		&api.CreateLoggerRequest{
			Topic:       topic,
			Description: description,
		}); err != nil {
		return
	}
	return
}

// UpdateLogger 更新日志主题
// @key 主题标识
// @description 描述
// topic 与 description 为空同样会更新
func (x *Transfer) UpdateLogger(ctx context.Context, key string, description string) (err error) {
	if _, err = x.client.UpdateLogger(ctx,
		&api.UpdateLoggerRequest{
			Key:         key,
			Description: description,
		}); err != nil {
		return
	}
	return
}

// DeleteLogger 删除日志主题
// @key 主题标识
func (x *Transfer) DeleteLogger(ctx context.Context, key string) (err error) {
	if _, err = x.client.DeleteLogger(ctx, &api.DeleteLoggerRequest{Key: key}); err != nil {
		return
	}
	return
}

// Info 获取日志主题详情
// @key 主题标识
func (x *Transfer) Info(ctx context.Context, key string) (*api.InfoReply, error) {
	return x.client.Info(ctx, &api.InfoRequest{Key: key})
}

// Publish 投递日志
// @topic 主题
// @data 内容
func (x *Transfer) Publish(ctx context.Context, topic string, data interface{}) (err error) {
	var payload []byte
	if payload, err = msgpack.Marshal(data); err != nil {
		return
	}
	if _, err = x.client.Publish(ctx,
		&api.PublishRequest{
			Topic:   topic,
			Payload: payload,
		}); err != nil {
		return
	}
	return
}
