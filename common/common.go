package common

import (
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/mongo"
)

type Inject struct {
	Values *Values
	Db     *mongo.Database
	Nats   *nats.Conn
	Js     nats.JetStreamContext
}

type Values struct {
	Address   string   `yaml:"address"`
	Namespace string   `yaml:"namespace"`
	Database  Database `yaml:"database"`
	Nats      Nats     `yaml:"nats"`
}

type Database struct {
	Uri        string `yaml:"uri"`
	Name       string `yaml:"name"`
	Collection string `yaml:"collection"`
}

type Nats struct {
	Hosts []string `yaml:"hosts"`
	Nkey  string   `yaml:"nkey"`
}
