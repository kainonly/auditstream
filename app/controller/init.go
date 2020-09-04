package controller

import (
	"elastic-transfer/app/manage"
	pb "elastic-transfer/router"
)

type controller struct {
	pb.UnimplementedRouterServer
	manager *manage.ElasticManager
}

func New(manager *manage.ElasticManager) *controller {
	c := new(controller)
	c.manager = manager
	return c
}
