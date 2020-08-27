package controller

import (
	"context"
	pb "microtools-gossh/router"
)

func (c *controller) Exec(ctx context.Context, param *pb.ExecParameter) (*pb.ExecResponse, error) {
	output, err := c.manager.Exec(param.Identity, param.Bash)
	if err != nil {
		return c.execErrorResponse(err)
	}
	return c.execSuccessResponse(output)
}

func (c *controller) execErrorResponse(err error) (*pb.ExecResponse, error) {
	return &pb.ExecResponse{
		Error: 1,
		Msg:   err.Error(),
	}, nil
}

func (c *controller) execSuccessResponse(data []byte) (*pb.ExecResponse, error) {
	return &pb.ExecResponse{
		Error: 0,
		Msg:   "ok",
		Data:  string(data),
	}, nil
}
