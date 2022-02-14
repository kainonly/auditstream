package api

import (
	"context"
	"github.com/google/wire"
	zlogging "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/nats-io/nats.go"
	"github.com/weplanx/transfer/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var Provides = wire.NewSet(
	New,
)

func New(x *common.Inject) (s *grpc.Server, err error) {
	// 初始化命名空间
	if _, err = x.Js.AddStream(&nats.StreamConfig{
		Name:     "namespaces",
		Subjects: []string{"namespaces.>"},
		MaxMsgs:  1,
	}); err != nil {
		return
	}

	// 存储索引
	if _, err = x.Db.Collection(x.Values.Database.Collection).Indexes().
		CreateMany(context.TODO(), []mongo.IndexModel{
			{
				Keys: bson.M{"key": 1},
				Options: options.Index().
					SetName("uk_key").
					SetUnique(true),
			},
			{
				Keys: bson.M{"topic": 1},
				Options: options.Index().
					SetName("uk_topic").
					SetUnique(true),
			},
		}); err != nil {
		return
	}

	// 中间件
	var zlogger *zap.Logger
	if zlogger, err = zap.NewProduction(); err != nil {
		return
	}
	zlogging.ReplaceGrpcLoggerV2(zlogger)
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			zlogging.UnaryServerInterceptor(zlogger),
			recovery.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			zlogging.StreamServerInterceptor(zlogger),
			recovery.StreamServerInterceptor(),
		),
	}

	// 设置 TLS
	if x.Values.TLS.Cert != "" && x.Values.TLS.Key != "" {
		var creds credentials.TransportCredentials
		if creds, err = credentials.NewServerTLSFromFile(
			x.Values.TLS.Cert,
			x.Values.TLS.Key,
		); err != nil {
			return
		}
		opts = append(opts, grpc.Creds(creds))
	}

	s = grpc.NewServer(opts...)
	RegisterAPIServer(s, &API{Inject: x})
	return
}
