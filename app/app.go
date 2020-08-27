package app

import (
	"google.golang.org/grpc"
	"microtools-gossh/app/controller"
	"microtools-gossh/app/manage"
	"microtools-gossh/app/types"
	pb "microtools-gossh/router"
	"net"
	"net/http"
	_ "net/http/pprof"
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
	manager, err := manage.NewClientManager()
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
