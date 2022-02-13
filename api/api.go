package api

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/thoas/go-funk"
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

// CollName 集合名称
// 默认 logger
func (x *API) CollName() string {
	if x.Values.Database.Collection == "" {
		return "logger"
	}
	return x.Values.Database.Collection
}

// ExistsStream 判断流命名是否存在
// 不能包含空格、制表符、句点 (.)、大于 (>) 或星号 (*)
func (x *API) ExistsStream(name string) bool {
	var names []string
	for x := range x.Js.StreamNames() {
		names = append(names, x)
	}
	return funk.Contains(names, name)
}

// GetLoggers 获取日志主题设置
func (x *API) GetLoggers(ctx context.Context, _ *empty.Empty) (rep *GetLoggersReply, err error) {
	var cursor *mongo.Cursor
	if cursor, err = x.Db.Collection(x.CollName()).Find(ctx, bson.M{}); err != nil {
		return
	}
	rep = new(GetLoggersReply)
	if err = cursor.All(ctx, &rep.Data); err != nil {
		return
	}
	return
}

// CreateLogger 创建日志主题
func (x *API) CreateLogger(ctx context.Context, req *CreateLoggerRequest) (_ *empty.Empty, err error) {
	key := uuid.New().String()
	var session mongo.Session
	if session, err = x.Mongo.StartSession(); err != nil {
		return
	}
	defer session.EndSession(ctx)
	if _, err = session.WithTransaction(ctx,
		func(sessCtx mongo.SessionContext) (_ interface{}, err error) {
			wcMajority := writeconcern.New(writeconcern.WMajority(), writeconcern.WTimeout(time.Second))
			wcMajorityCollectionOpts := options.Collection().SetWriteConcern(wcMajority)
			if _, err = x.Db.Collection(x.CollName(), wcMajorityCollectionOpts).
				InsertOne(sessCtx, bson.M{
					"key":         key,
					"topic":       req.Topic,
					"description": req.Description,
				}); err != nil {
				return
			}
			subject := fmt.Sprintf(`%s.%s`, x.Values.Namespace, req.Topic)
			if _, err = x.Js.AddStream(&nats.StreamConfig{
				Name:        key,
				Subjects:    []string{subject},
				Description: req.Description,
				Retention:   nats.WorkQueuePolicy,
			}); err != nil {
				return
			}
			return
		},
	); err != nil {
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
			wcMajority := writeconcern.New(writeconcern.WMajority(), writeconcern.WTimeout(time.Second))
			wcMajorityCollectionOpts := options.Collection().SetWriteConcern(wcMajority)
			if _, err = x.Db.Collection(x.CollName(), wcMajorityCollectionOpts).
				UpdateOne(sessCtx, bson.M{"key": req.Key}, bson.M{
					"$set": bson.M{
						"topic":       req.Topic,
						"description": req.Description,
					},
				}); err != nil {
				return
			}
			subject := fmt.Sprintf(`%s.%s`, x.Values.Namespace, req.Topic)
			if _, err = x.Js.UpdateStream(&nats.StreamConfig{
				Name:        req.Key,
				Subjects:    []string{subject},
				Description: req.Description,
				Retention:   nats.WorkQueuePolicy,
			}); err != nil {
				return
			}
			return
		},
	); err != nil {
		return
	}
	return &empty.Empty{}, nil
}

// DeleteLogger 删除日志主题
func (x *API) DeleteLogger(ctx context.Context, req *DeleteLoggerRequest) (_ *empty.Empty, err error) {
	if _, err = x.Db.Collection(x.CollName()).
		DeleteOne(ctx, bson.M{"key": req.Key}); err != nil {
		return
	}
	if x.ExistsStream(req.Key) {
		if err = x.Js.DeleteStream(req.Key); err != nil {
			return
		}
	}
	return &empty.Empty{}, nil
}

// Info 获取详情
func (x *API) Info(ctx context.Context, req *InfoRequest) (rep *InfoReply, err error) {
	var info *nats.StreamInfo
	if info, err = x.Js.StreamInfo(req.Key); err != nil {
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
	subject := fmt.Sprintf(`%s.%s`, x.Values.Namespace, req.Topic)
	if err = x.Nats.Publish(subject, req.Payload); err != nil {
		return
	}

	return &empty.Empty{}, nil
}
