package app

import (
	"elastic-transfer/app/controller"
	"elastic-transfer/app/manage"
	"elastic-transfer/app/mq"
	"elastic-transfer/app/types"
	pb "elastic-transfer/router"
	"google.golang.org/grpc"
	"net"
	"net/http"
)

type App struct {
	option *types.Config
}

func New(config types.Config) *App {
	app := new(App)
	app.option = &config
	return app
}

func (app *App) Start() (err error) {
	// Turn on debugging
	if app.option.Debug {
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}
	// Start microservice
	listen, err := net.Listen("tcp", app.option.Listen)
	if err != nil {
		return
	}
	server := grpc.NewServer()
	mqlib, err := mq.NewMessageQueue(app.option.Mq)
	if err != nil {
		return
	}
	manager, err := manage.NewElasticManager(
		app.option.Elastic,
		mqlib,
	)
	if err != nil {
		return
	}
	pb.RegisterRouterServer(
		server,
		controller.New(manager),
	)
	server.Serve(listen)
	return
}
