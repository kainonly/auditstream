package api

import (
	"context"
	"fmt"
	"github.com/google/wire"
	zlogging "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/nats-io/nats.go"
	"github.com/thoas/go-funk"
	"github.com/weplanx/transfer/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var Provides = wire.NewSet(New)

func New(i *common.Inject) (s *grpc.Server, err error) {
	// 初始化命名空间
	if _, err = i.Js.AddStream(&nats.StreamConfig{
		Name:     "namespaces",
		Subjects: []string{"namespaces.*.ready"},
		MaxMsgs:  1,
	}); err != nil {
		return
	}
	if _, err = i.Js.AddStream(&nats.StreamConfig{
		Name:      "namespaces-event",
		Subjects:  []string{"namespaces.*.event"},
		Retention: nats.InterestPolicy,
	}); err != nil {
		return
	}

	// 存储索引
	ctx := context.Background()
	if _, err = i.Db.Collection(i.Values.Database.Collection).Indexes().
		CreateMany(ctx, []mongo.IndexModel{
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
	if i.Values.TLS.Cert != "" && i.Values.TLS.Key != "" {
		var creds credentials.TransportCredentials
		if creds, err = credentials.NewServerTLSFromFile(
			i.Values.TLS.Cert,
			i.Values.TLS.Key,
		); err != nil {
			return
		}
		opts = append(opts, grpc.Creds(creds))
	}

	s = grpc.NewServer(opts...)
	api := &API{Inject: i}

	// 检测是否同步
	streams := api.streams()
	loggers := make([]*Logger, 0)
	if err = api.loggers(ctx, &loggers); err != nil {
		return
	}
	notExists := funk.Filter(loggers, func(v *Logger) bool {
		return !funk.Contains(streams, v.Key)
	})
	for _, v := range notExists.([]*Logger) {
		subject := fmt.Sprintf(`%s.%s`, i.Values.Namespace, v.Topic)
		if _, err = i.Js.AddStream(&nats.StreamConfig{
			Name:        v.Key,
			Subjects:    []string{subject},
			Description: v.Topic,
			Retention:   nats.WorkQueuePolicy,
		}); err != nil {
			return
		}
	}

	// 发布初始化主题
	topics := make([]string, len(loggers))
	for i, v := range loggers {
		topics[i] = v.Topic
	}
	if err = api.ready(topics); err != nil {
		return
	}

	RegisterAPIServer(s, api)
	return
}
