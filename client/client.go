package client

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
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

func (x *Transfer) GetLoggers(ctx context.Context) (result []*api.Logger, err error) {
	var rep *api.GetLoggersReply
	if rep, err = x.client.GetLoggers(ctx, &empty.Empty{}); err != nil {
		return
	}
	return rep.GetData(), nil
}

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

func (x *Transfer) UpdateLogger(ctx context.Context, key string, topic string, description string) (err error) {
	if _, err = x.client.UpdateLogger(ctx,
		&api.UpdateLoggerRequest{
			Key:         key,
			Topic:       topic,
			Description: description,
		}); err != nil {
		return
	}
	return
}

func (x *Transfer) DeleteLogger(ctx context.Context, key string) (err error) {
	if _, err = x.client.DeleteLogger(ctx, &api.DeleteLoggerRequest{Key: key}); err != nil {
		return
	}
	return
}

func (x *Transfer) Info(ctx context.Context, key string) (*api.InfoReply, error) {
	return x.client.Info(ctx, &api.InfoRequest{Key: key})
}

func (x *Transfer) Publish(ctx context.Context, topic string, payload []byte) (err error) {
	if _, err = x.client.Publish(ctx,
		&api.PublishRequest{
			Topic:   topic,
			Payload: payload,
		}); err != nil {
		return
	}
	return
}
