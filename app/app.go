package app

import (
	"elastic-transfer/app/controller"
	"elastic-transfer/app/manage"
	"elastic-transfer/app/mq"
	"elastic-transfer/app/types"
	pb "elastic-transfer/router"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/panjf2000/ants/v2"
	"google.golang.org/grpc"
	"net"
	"net/http"
	_ "net/http/pprof"
)

func Application(option *types.Config) (err error) {
	// Turn on debugging
	if option.Debug != "" {
		go func() {
			http.ListenAndServe(option.Debug, nil)
		}()
	}
	// Start microservice
	listen, err := net.Listen("tcp", option.Listen)
	if err != nil {
		return
	}
	server := grpc.NewServer()
	elastic, err := elasticsearch.NewClient(option.Elastic)
	if err != nil {
		return
	}
	mqclient, err := mq.NewMessageQueue(option.Mq)
	if err != nil {
		return
	}
	pool, err := ants.NewPool(10000, ants.WithPreAlloc(true))
	if err != nil {
		return
	}
	defer pool.Release()
	manager, err := manage.NewElasticManager(
		elastic,
		mqclient,
		pool,
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
