package main

import (
	"github.com/smallnest/rpcx/server"
	"github.com/weplanx/transfer/api"
)

func main() {
	s := server.NewServer()
	s.RegisterName("API", new(api.API), "")
	s.Serve("tcp", ":8972")
}
