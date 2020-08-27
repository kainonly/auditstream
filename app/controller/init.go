package controller

import (
	"microtools-gossh/app/manage"
	pb "microtools-gossh/router"
)

type controller struct {
	pb.UnimplementedRouterServer
	manager *manage.ClientManager
}

func New(manager *manage.ClientManager) *controller {
	c := new(controller)
	c.manager = manager
	return c
}

func (c *controller) response(err error) (*pb.Response, error) {
	if err != nil {
		return &pb.Response{
			Error: 1,
			Msg:   err.Error(),
		}, nil
	} else {
		return &pb.Response{
			Error: 0,
			Msg:   "ok",
		}, nil
	}
}
