package app

import (
	"elastic-transfer/app/controller"
	"elastic-transfer/app/manage"
	"elastic-transfer/app/mq"
	"elastic-transfer/app/types"
	pb "elastic-transfer/router"
	"github.com/elastic/go-elasticsearch/v8"
	"google.golang.org/grpc"
	"net"
	"net/http"
	_ "net/http/pprof"
)

func Application(option *types.Config) (err error) {
	// Turn on debugging
	if option.Debug {
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}
	// Start microservice
	listen, err := net.Listen("tcp", option.Listen)
	if err != nil {
		return
	}
	server := grpc.NewServer()
	elastic, err := elasticsearch.NewClient(option.Elastic)
	mqclient, err := mq.NewMessageQueue(option.Mq)
	if err != nil {
		return
	}
	manager, err := manage.NewElasticManager(
		elastic,
		mqclient,
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
