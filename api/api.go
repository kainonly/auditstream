package api

import "context"

type API struct{}

type Reply struct {
	Code int64
	Msg  string
}

type SendArgs struct {
	Code int64
}

func (t *API) Send(ctx context.Context, args *SendArgs, reply *Reply) error {
	reply.Code = args.Code
	reply.Msg = "OK"
	return nil
}
