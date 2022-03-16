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

// Publish 投递日志
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
