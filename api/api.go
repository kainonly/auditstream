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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

// Logger 获取日志主题设置
func (x *API) Logger(ctx context.Context, _ *empty.Empty) (rep *LoggerReply, err error) {
	var cursor *mongo.Cursor
	if cursor, err = x.Db.Collection(x.CollName()).Find(ctx, bson.M{}); err != nil {
		return
	}
	if err = cursor.All(ctx, &rep.Data); err != nil {
		return
	}
	return
}

// CreateLogger 创建日志主题
func (x *API) CreateLogger(ctx context.Context, req *CreateLoggerRequest) (_ *empty.Empty, err error) {
	key := uuid.New().String()
	if _, err = x.Db.Collection(x.CollName()).
		InsertOne(ctx, Logger{
			Key:         key,
			Topic:       req.Topic,
			Description: req.Description,
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
}

// DeleteLogger 删除日志主题
func (x *API) DeleteLogger(ctx context.Context, req *DeleteLoggerRequest) (_ *empty.Empty, err error) {
	var data Logger
	var oid primitive.ObjectID
	if oid, err = primitive.ObjectIDFromHex(req.Id); err != nil {
		return
	}
	if err = x.Db.Collection(x.CollName()).
		FindOne(ctx, bson.M{"_id": oid}).Decode(&data); err != nil {
		return
	}
	if err = x.Js.DeleteStream(data.Key); err != nil {
		return
	}
	if _, err = x.Db.Collection(x.CollName()).
		DeleteOne(ctx, bson.M{"_id": req.Id}); err != nil {
		return
	}
	return
}

// Publish 投递日志
func (x *API) Publish(ctx context.Context, req *PublishRequest) (_ *empty.Empty, err error) {
	subject := fmt.Sprintf(`%s.%s`, x.Values.Namespace, req.Topic)
	if err = x.Nats.Publish(subject, req.Payload); err != nil {
		return
	}
	return
}
