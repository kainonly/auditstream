package app

import (
	"google.golang.org/grpc"
	"net"
	"net/http"
	"transfer-microservice/app/types"
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
	//manager, err := manage.NewElasticManager(app.option.Elastic)
	//if err != nil {
	//	return
	//}
	server.Serve(listen)
	return
}
