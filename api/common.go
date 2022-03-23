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
	readyName := fmt.Sprintf(`%s:logs`, namespace)
	readySubject := fmt.Sprintf(`%s.logs`, namespace)
	if _, err = i.Js.AddStream(&nats.StreamConfig{
		Name:     readyName,
		Subjects: []string{readySubject},
		MaxMsgs:  1,
	}, nats.Context(ctx)); err != nil {
		return
	}
	eventsName := fmt.Sprintf(`%s:logs:events`, namespace)
	eventsSubject := fmt.Sprintf(`%s.logs.events`, namespace)
	if _, err = i.Js.AddStream(&nats.StreamConfig{
		Name:      eventsName,
		Subjects:  []string{eventsSubject},
		Retention: nats.InterestPolicy,
	}, nats.Context(ctx)); err != nil {
		return
	}
	i.Log.Info("已初始化状态流",
		zap.Any("state", []interface{}{readySubject, eventsSubject}),
	)

	// 初始化存储索引
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
	i.Log.Info("已初始化存储索引",
		zap.Any("state", []interface{}{"uk_key", "uk_topic"}),
	)

	// 中间件
	//zlogging.ReplaceGrpcLoggerV2(i.Log)
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			zlogging.UnaryServerInterceptor(i.Log),
			recovery.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			zlogging.StreamServerInterceptor(i.Log),
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
	loggers := make([]*Transfer, 0)
	if err = api.getValues(ctx, &loggers); err != nil {
		return
	}
	notExists := funk.Filter(loggers, func(v *Transfer) bool {
		return !funk.Contains(streams, fmt.Sprintf(`%s:logs:%s`, namespace, v.Key))
	})
	for _, v := range notExists.([]*Transfer) {
		name := fmt.Sprintf(`%s:logs:%s`, namespace, v.Key)
		subject := fmt.Sprintf(`%s.logs.%s`, namespace, v.Topic)
		if _, err = i.Js.AddStream(&nats.StreamConfig{
			Name:        name,
			Subjects:    []string{subject},
			Description: v.Topic,
			Retention:   nats.WorkQueuePolicy,
		}, nats.Context(ctx)); err != nil {
			return
		}
		i.Log.Info("同步配置",
			zap.String("name", name),
			zap.String("subject", subject),
		)
	}

	// 发布日志主题配置
	if err = api.publishReady(ctx, loggers); err != nil {
		return
	}

	RegisterAPIServer(s, api)
	return
}
