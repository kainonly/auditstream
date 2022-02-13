package app

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/weplanx/transfer/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type API struct {
	Values *common.Values
	Db     *mongo.Database
	Nats   *nats.Conn
	Js     nats.JetStreamContext
}

func (x *API) name() string {
	if x.Values.Database.Collection == "" {
		return "logger"
	}
	return x.Values.Database.Collection
}

type LoggerReply struct {
	Data []map[string]interface{}
}

func (x *API) Logger(ctx context.Context, _ *Empty, rep *LoggerReply) (err error) {
	var cursor *mongo.Cursor
	if cursor, err = x.Db.Collection(x.name()).Find(ctx, bson.M{}); err != nil {
		return
	}
	data := make([]map[string]interface{}, 0)
	if err = cursor.All(ctx, &data); err != nil {
		return
	}
	rep = &LoggerReply{Data: data}
	return
}

type CreateLoggerRequest struct {
	Topic       string
	Description string
}

func (x *API) CreateLogger(ctx context.Context, req *CreateLoggerRequest, _ *Empty) (err error) {
	key := uuid.New().String()
	if _, err = x.Db.Collection(x.name()).
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

type DeleteLoggerRequest struct {
	Id primitive.ObjectID
}

func (x *API) DeleteLogger(ctx context.Context, req *DeleteLoggerRequest, _ *Empty) (err error) {
	var data Logger
	if err = x.Db.Collection(x.name()).
		FindOne(ctx, bson.M{"_id": req.Id}).Decode(&data); err != nil {
		return
	}
	if err = x.Js.DeleteStream(data.Key); err != nil {
		return
	}
	if _, err = x.Db.Collection(x.name()).
		DeleteOne(ctx, bson.M{"_id": req.Id}); err != nil {
		return
	}
	return
}

type PublishRequest struct {
	Topic   string
	Payload []byte
}

func (x *API) Publish(_ context.Context, req *PublishRequest, _ *Empty) error {
	subject := fmt.Sprintf(`%s.%s`, x.Values.Namespace, req.Topic)
	return x.Nats.Publish(subject, req.Payload)
}
