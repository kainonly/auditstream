package main

import (
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"ssh-microservice/client"
	"ssh-microservice/common"
	"ssh-microservice/controller"
	pb "ssh-microservice/router"
)

func main() {
	if _, err := os.Stat("./config/autoload"); os.IsNotExist(err) {
		os.Mkdir("./config/autoload", os.ModeDir)
	}
	in, err := ioutil.ReadFile("./config/config.yml")
	if err != nil {
		log.Fatalln(err)
	}
	cfg := common.AppOption{}
	err = yaml.Unmarshal(in, &cfg)
	if err != nil {
		log.Fatalln(err)
	}
	if cfg.Debug {
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}
	common.InitBufPool(cfg.Pool)
	cli := client.Create()
	listen, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		log.Fatalln(err)
	}
	server := grpc.NewServer()
	pb.RegisterRouterServer(
		server,
		controller.New(cli),
	)
	server.Serve(listen)
}
