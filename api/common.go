package api

import (
	"github.com/google/wire"
	"github.com/weplanx/transfer/common"
	"google.golang.org/grpc"
)

var Provides = wire.NewSet(
	New,
)

func New(i *common.Inject) (s *grpc.Server, err error) {
	s = grpc.NewServer()
	RegisterAPIServer(s, &API{Inject: i})
	return
}
