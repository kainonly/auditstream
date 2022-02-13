package app

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/google/wire"
	"github.com/smallnest/rpcx/server"
	"github.com/weplanx/transfer/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
)

var Provides = wire.NewSet(
	wire.Struct(new(API), "*"),
	New,
)

func New(
	values *common.Values,
	db *mongo.Database,
	api *API,
) (s *server.Server, err error) {
	var caCertPEM []byte
	if caCertPEM, err = ioutil.ReadFile(values.TLS.Ca); err != nil {
		return
	}
	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM(caCertPEM); !ok {
		return nil, common.CertsFromPEMError
	}
	var cert tls.Certificate
	if cert, err = tls.LoadX509KeyPair(
		values.TLS.Cert,
		values.TLS.Key,
	); err != nil {
		return
	}
	config := &tls.Config{
		RootCAs:      roots,
		Certificates: []tls.Certificate{cert},
	}
	s = server.NewServer(server.WithTLSConfig(config))
	if err = s.RegisterName("API", api, ""); err != nil {
		return
	}
	if _, err = db.Collection(values.Database.Collection).Indexes().
		CreateMany(context.TODO(), []mongo.IndexModel{
			{
				Keys:    bson.M{"key": 1},
				Options: options.Index().SetName("uk_key").SetUnique(true),
			},
			{
				Keys:    bson.M{"topic": 1},
				Options: options.Index().SetName("uk_topic").SetUnique(true),
			},
		}); err != nil {
		return
	}
	return
}

type Logger struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Key         string             `bson:"key"`
	Topic       string             `bson:"topic"`
	Description string             `bson:"description"`
}

type Empty struct{}
