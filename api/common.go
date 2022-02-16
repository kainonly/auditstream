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
	ctx := context.Background()
	namespace := i.Values.Namespace

	// 初始化状态流
	if _, err = i.Js.AddStream(&nats.StreamConfig{
		Name: fmt.Sprintf(`namespaces:%s`, namespace),
		Subjects: []string{
			fmt.Sprintf(`namespaces.%s.ready`, namespace),
		},
		MaxMsgs: 1,
	}, nats.Context(ctx)); err != nil {
		return
	}
	if _, err = i.Js.AddStream(&nats.StreamConfig{
		Name: fmt.Sprintf(`namespaces:%s:event`, namespace),
		Subjects: []string{
			fmt.Sprintf(`namespaces.%s.event`, namespace),
		},
		Retention: nats.InterestPolicy,
	}, nats.Context(ctx)); err != nil {
		return
	}

	// 存储索引
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
		tls := i.Values.TLS
		var creds credentials.TransportCredentials
		if creds, err = credentials.NewServerTLSFromFile(
			tls.Cert,
			tls.Key,
		); err != nil {
			return
		}
		opts = append(opts, grpc.Creds(creds))
	}

	s = grpc.NewServer(opts...)
	api := &API{Inject: i}

	// 检测数据流是否同步
	streams := api.streams()
	loggers := make([]*Logger, 0)
	if err = api.loggers(ctx, &loggers); err != nil {
		return
	}
	notExists := funk.Filter(loggers, func(v *Logger) bool {
		return !funk.Contains(streams, v.Key)
	})
	for _, v := range notExists.([]*Logger) {
		name := fmt.Sprintf(`namespaces:%s:%s`, namespace, v.Key)
		subject := fmt.Sprintf(`%s.%s`, namespace, v.Topic)
		if _, err = i.Js.AddStream(&nats.StreamConfig{
			Name:        name,
			Subjects:    []string{subject},
			Description: v.Topic,
			Retention:   nats.WorkQueuePolicy,
		}, nats.Context(ctx)); err != nil {
			return
		}
	}

	// 发布日志主题配置
	if err = api.ready(ctx, loggers); err != nil {
		return
	}

	RegisterAPIServer(s, api)
	return
}
