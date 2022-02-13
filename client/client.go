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

type LoggerInfo struct {
	Key         string
	Topic       string
	Description string
}

func (x *Transfer) Logger(ctx context.Context) (result []LoggerInfo, err error) {
	var rep *api.LoggerReply
	if rep, err = x.client.Logger(ctx, &empty.Empty{}); err != nil {
		return
	}
	result = make([]LoggerInfo, len(rep.GetData()))
	for i, v := range rep.GetData() {
		result[i] = LoggerInfo{
			Key:         v.Key,
			Topic:       v.Topic,
			Description: v.Description,
		}
	}
	return
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

func (x *Transfer) DeleteLogger(ctx context.Context, key string) (err error) {
	if _, err = x.client.DeleteLogger(ctx, &api.DeleteLoggerRequest{Key: key}); err != nil {
		return
	}
	return
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
