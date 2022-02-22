package api

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/nats-io/nats.go"
	"github.com/thoas/go-funk"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/weplanx/transfer/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"time"
)

type API struct {
	UnimplementedAPIServer
	*common.Inject
}

// 集合名称，默认 logger
func (x *API) name() string {
	if x.Values.Database.Collection == "" {
		return "logger"
	}
	return x.Values.Database.Collection
}

// 获取所有流
func (x *API) streams() (names []string) {
	for x := range x.Js.StreamNames() {
		names = append(names, x)
	}
	return
}

// 获取日志主题
func (x *API) loggers(ctx context.Context, v interface{}) (err error) {
	var cursor *mongo.Cursor
	if cursor, err = x.Db.Collection(x.name()).Find(ctx, bson.M{}); err != nil {
		return
	}
	if err = cursor.All(ctx, v); err != nil {
		return
	}
	return
}

// 发布初始化主题
func (x *API) ready(ctx context.Context, loggers []*Logger) (err error) {
	topics := make([]string, len(loggers))
	for i, v := range loggers {
		topics[i] = v.Topic
	}
	var data []byte
	if data, err = msgpack.Marshal(topics); err != nil {
		return
	}
	subject := fmt.Sprintf(`logs.%s.ready`, x.Values.Namespace)
	if _, err = x.Js.Publish(subject, data, nats.Context(ctx)); err != nil {
		return
	}
	return
}

// 事件状态
func (x *API) event(ctx context.Context, topic string, action string) (err error) {
	var data []byte
	if data, err = msgpack.Marshal(map[string]string{
		"topic":  topic,
		"action": action,
	}); err != nil {
		return
	}
	subject := fmt.Sprintf(`logs.%s.event`, x.Values.Namespace)
	if _, err = x.Js.Publish(subject, data, nats.Context(ctx)); err != nil {
		return
	}
	return
}

// 变更状态配置
func (x *API) changed(ctx context.Context) (err error) {
	loggers := make([]*Logger, 0)
	if err = x.loggers(ctx, &loggers); err != nil {
		return
	}
	if err = x.ready(ctx, loggers); err != nil {
		return
	}
	return
}

// GetLoggers 获取日志主题设置
func (x *API) GetLoggers(ctx context.Context, _ *empty.Empty) (rep *GetLoggersReply, err error) {
	rep = new(GetLoggersReply)
	if err = x.loggers(ctx, &rep.Data); err != nil {
		return
	}
	return
}

// CreateLogger 创建日志主题
func (x *API) CreateLogger(ctx context.Context, req *CreateLoggerRequest) (_ *empty.Empty, err error) {
	var session mongo.Session
	if session, err = x.Mongo.StartSession(); err != nil {
		return
	}
	defer session.EndSession(ctx)
	if _, err = session.WithTransaction(ctx,
		func(sessCtx mongo.SessionContext) (_ interface{}, err error) {
			wcMajority := writeconcern.New(writeconcern.WMajority(), writeconcern.WTimeout(time.Second))
			wcMajorityCollectionOpts := options.Collection().SetWriteConcern(wcMajority)
			if _, err = x.Db.Collection(x.name(), wcMajorityCollectionOpts).
				InsertOne(sessCtx, bson.M{
					"key":         req.Key,
					"topic":       req.Topic,
					"description": req.Description,
				}); err != nil {
				return
			}
			name := fmt.Sprintf(`logs:%s:%s`, x.Values.Namespace, req.Key)
			subject := fmt.Sprintf(`logs.%s.%s`, x.Values.Namespace, req.Topic)
			if _, err = x.Js.AddStream(&nats.StreamConfig{
				Name:        name,
				Subjects:    []string{subject},
				Description: req.Description,
				Retention:   nats.WorkQueuePolicy,
			}); err != nil {
				return
			}
			if err = x.event(ctx, req.Topic, "create"); err != nil {
				return
			}
			return
		},
	); err != nil {
		return
	}
	if err = x.changed(ctx); err != nil {
		return
	}
	return &empty.Empty{}, nil
}

// UpdateLogger 更新日志主题
func (x *API) UpdateLogger(ctx context.Context, req *UpdateLoggerRequest) (_ *empty.Empty, err error) {
	var session mongo.Session
	if session, err = x.Mongo.StartSession(); err != nil {
		return
	}
	defer session.EndSession(ctx)
	if _, err = session.WithTransaction(ctx,
		func(sessCtx mongo.SessionContext) (_ interface{}, err error) {
			var data Logger
			if err = x.Db.Collection(x.name()).
				FindOne(sessCtx, bson.M{"key": req.Key}).
				Decode(&data); err != nil {
				return
			}

			wcMajority := writeconcern.New(writeconcern.WMajority(), writeconcern.WTimeout(time.Second))
			wcMajorityCollectionOpts := options.Collection().SetWriteConcern(wcMajority)
			if _, err = x.Db.Collection(x.name(), wcMajorityCollectionOpts).
				UpdateOne(sessCtx, bson.M{"key": req.Key}, bson.M{
					"$set": bson.M{
						"description": req.Description,
					},
				}); err != nil {
				return
			}
			name := fmt.Sprintf(`logs:%s:%s`, x.Values.Namespace, req.Key)
			subject := fmt.Sprintf(`logs.%s.%s`, x.Values.Namespace, data.Topic)
			if _, err = x.Js.UpdateStream(&nats.StreamConfig{
				Name:        name,
				Subjects:    []string{subject},
				Description: req.Description,
				Retention:   nats.WorkQueuePolicy,
			}, nats.Context(ctx)); err != nil {
				return
			}

			return
		},
	); err != nil {
		return
	}
	if err = x.changed(ctx); err != nil {
		return
	}
	return &empty.Empty{}, nil
}

// DeleteLogger 删除日志主题
func (x *API) DeleteLogger(ctx context.Context, req *DeleteLoggerRequest) (_ *empty.Empty, err error) {
	var data Logger
	if err = x.Db.Collection(x.name()).
		FindOne(ctx, bson.M{"key": req.Key}).
		Decode(&data); err != nil {
		return
	}
	if _, err = x.Db.Collection(x.name()).
		DeleteOne(ctx, bson.M{"key": req.Key}); err != nil {
		return
	}

	name := fmt.Sprintf(`logs:%s:%s`, x.Values.Namespace, req.Key)
	if funk.Contains(x.streams(), name) {
		if err = x.Js.DeleteStream(name, nats.Context(ctx)); err != nil {
			return
		}
		if err = x.event(ctx, data.Topic, "delete"); err != nil {
			return
		}
	}

	if err = x.changed(ctx); err != nil {
		return
	}
	return &empty.Empty{}, nil
}

// Info 获取详情
func (x *API) Info(ctx context.Context, req *InfoRequest) (rep *InfoReply, err error) {
	var info *nats.StreamInfo
	name := fmt.Sprintf(`logs:%s:%s`, x.Values.Namespace, req.Key)
	if info, err = x.Js.StreamInfo(name, nats.Context(ctx)); err != nil {
		return
	}
	return &InfoReply{
		State: &InfoState{
			Messages:      info.State.Msgs,
			Bytes:         info.State.Bytes,
			FirstSeq:      info.State.FirstSeq,
			FirstTime:     info.State.FirstTime.Unix(),
			LastSeq:       info.State.LastSeq,
			LastTime:      info.State.FirstTime.Unix(),
			ConsumerCount: int64(info.State.Consumers),
		},
		CreateTime: info.Created.Unix(),
	}, nil
}

// Publish 投递日志
func (x *API) Publish(ctx context.Context, req *PublishRequest) (_ *empty.Empty, err error) {
	subject := fmt.Sprintf(`logs.%s.%s`, x.Values.Namespace, req.Topic)
	if _, err = x.Js.Publish(subject, req.Payload, nats.Context(ctx)); err != nil {
		return
	}
	return &empty.Empty{}, nil
}
