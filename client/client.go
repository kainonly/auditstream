package client

import (
	"github.com/weplanx/transfer/api"
	"google.golang.org/grpc"
)

type Transfer struct {
	client api.APIClient
	conn   *grpc.ClientConn
}

func New(addr string) (x *Transfer, err error) {
	x = new(Transfer)
	if x.conn, err = grpc.Dial(addr); err != nil {
		return
	}
	x.client = api.NewAPIClient(x.conn)
	return
}

func (x *Transfer) Close() error {
	return x.conn.Close()
}

//func (x *Transfer) Logger(ctx context.Context) (reply *app.LoggerReply, err error) {
//	reply = new(app.LoggerReply)
//	if err = x.Client.Call(
//		context.Background(),
//		"Logger",
//		&app.Empty{},
//		reply,
//	); err != nil {
//		return
//	}
//	x.client.Logger(ctx, &empty.Empty{})
//	return
//}
//
//func (x *Transfer) CreateLogger(args app.CreateLoggerRequest) (err error) {
//	if err = x.Client.Call(
//		context.Background(),
//		"CreateLogger",
//		&args,
//		&app.Empty{},
//	); err != nil {
//		return
//	}
//	return
//}
//
//func (x *Transfer) DeleteLogger(args app.DeleteLoggerRequest) (err error) {
//	if err = x.Client.Call(
//		context.Background(),
//		"DeleteLogger",
//		&args,
//		&app.Empty{},
//	); err != nil {
//		return
//	}
//	return
//}
//
//func (x *Transfer) Publish(args app.PublishRequest) (err error) {
//	if err = x.Client.Call(
//		context.Background(),
//		"Publish",
//		&args,
//		&app.Empty{},
//	); err != nil {
//		return
//	}
//	return
//}
