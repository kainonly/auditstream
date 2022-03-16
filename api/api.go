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

// 集合名称
func (x *API) name() string {
	if x.Values.Database.Collection == "" {
		return "transfers"
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

// 获取配置数据
func (x *API) getValues(ctx context.Context, v interface{}) (err error) {
	var cursor *mongo.Cursor
	if cursor, err = x.Db.Collection(x.name()).Find(ctx, bson.M{}); err != nil {
		return
	}
	if err = cursor.All(ctx, v); err != nil {
		return
	}
	return
}

// 发布配置订阅
func (x *API) publishReady(ctx context.Context, transfers []*Transfer) (err error) {
	topics := make([]string, len(transfers))
	for k, v := range transfers {
		topics[k] = v.Topic
	}
	var data []byte
	if data, err = msgpack.Marshal(topics); err != nil {
		return
	}
	subject := fmt.Sprintf(`%s.logs`, x.Values.Namespace)
	if _, err = x.Js.Publish(subject, data, nats.Context(ctx)); err != nil {
		return
	}
	return
}

// 发布事件状态
func (x *API) publishEvent(ctx context.Context, topic string, action string) (err error) {
	var data []byte
	if data, err = msgpack.Marshal(map[string]string{
		"topic":  topic,
		"action": action,
	}); err != nil {
		return
	}
	subject := fmt.Sprintf(`%s.logs.events`, x.Values.Namespace)
	if _, err = x.Js.Publish(subject, data, nats.Context(ctx)); err != nil {
		return
	}
	return
}

// 变更状态配置
func (x *API) changed(ctx context.Context) (err error) {
	transfers := make([]*Transfer, 0)
	if err = x.getValues(ctx, &transfers); err != nil {
		return
	}
	if err = x.publishReady(ctx, transfers); err != nil {
		return
	}
	return
}

// Get 获取配置
func (x *API) Get(ctx context.Context, _ *empty.Empty) (rep *GetReply, err error) {
	rep = new(GetReply)
	if err = x.getValues(ctx, &rep.Data); err != nil {
		return
	}
	return
}

// Create 创建传输
func (x *API) Create(ctx context.Context, req *CreateRequest) (_ *empty.Empty, err error) {
	var count int64
	if count, err = x.Db.Collection(x.name()).
		CountDocuments(ctx, bson.M{"key": req.Key}); err != nil {
		return
	}
	if count != 0 {
		return &empty.Empty{}, nil
	}
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
			name := fmt.Sprintf(`%s:logs:%s`, x.Values.Namespace, req.Key)
			subject := fmt.Sprintf(`%s.logs.%s`, x.Values.Namespace, req.Topic)
			if _, err = x.Js.AddStream(&nats.StreamConfig{
				Name:        name,
				Subjects:    []string{subject},
				Description: req.Description,
				Retention:   nats.WorkQueuePolicy,
			}); err != nil {
				return
			}
			if err = x.publishEvent(ctx, req.Topic, "create"); err != nil {
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

// Update 更新传输
func (x *API) Update(ctx context.Context, req *UpdateRequest) (_ *empty.Empty, err error) {
	var session mongo.Session
	if session, err = x.Mongo.StartSession(); err != nil {
		return
	}
	defer session.EndSession(ctx)
	if _, err = session.WithTransaction(ctx,
		func(sessCtx mongo.SessionContext) (_ interface{}, err error) {
			var data Transfer
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
			name := fmt.Sprintf(`%s:logs:%s`, x.Values.Namespace, req.Key)
			subject := fmt.Sprintf(`%s.logs.%s`, x.Values.Namespace, data.Topic)
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

// Delete 删除传输
func (x *API) Delete(ctx context.Context, req *DeleteRequest) (_ *empty.Empty, err error) {
	var data Transfer
	if err = x.Db.Collection(x.name()).
		FindOneAndDelete(ctx, bson.M{"key": req.Key}).
		Decode(&data); err != nil {
		return
	}

	name := fmt.Sprintf(`%s:logs:%s`, x.Values.Namespace, req.Key)
	if funk.Contains(x.streams(), name) {
		if err = x.Js.DeleteStream(name, nats.Context(ctx)); err != nil {
			return
		}
		if err = x.publishEvent(ctx, data.Topic, "delete"); err != nil {
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
	name := fmt.Sprintf(`%s:logs:%s`, x.Values.Namespace, req.Key)
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
	subject := fmt.Sprintf(`%s.logs.%s`, x.Values.Namespace, req.Topic)
	if _, err = x.Js.Publish(subject, req.Payload, nats.Context(ctx)); err != nil {
		return
	}
	return &empty.Empty{}, nil
}
