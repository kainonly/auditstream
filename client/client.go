package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/smallnest/rpcx/client"
	"github.com/weplanx/transfer/app"
	"github.com/weplanx/transfer/common"
	"io/ioutil"
)

type Transfer struct {
	Client client.XClient
}

func New(addr string, TLSoption common.TLS) (x *Transfer, err error) {
	x = new(Transfer)
	var caCertPEM []byte
	if caCertPEM, err = ioutil.ReadFile(TLSoption.Ca); err != nil {
		return
	}
	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM(caCertPEM); !ok {
		return nil, common.CertsFromPEMError
	}
	option := client.DefaultOption
	option.TLSConfig = &tls.Config{
		RootCAs: roots,
	}
	var discovery *client.Peer2PeerDiscovery
	if discovery, err = client.NewPeer2PeerDiscovery(
		fmt.Sprintf(`quic@%s`, addr),
		"",
	); err != nil {
		return
	}
	x.Client = client.NewXClient("API",
		client.Failtry,
		client.RandomSelect,
		discovery,
		option,
	)
	return
}

func (x *Transfer) Close() error {
	return x.Client.Close()
}

func (x *Transfer) Logger() (reply *app.LoggerReply, err error) {
	reply = new(app.LoggerReply)
	if err = x.Client.Call(
		context.Background(),
		"Logger",
		&app.Empty{},
		reply,
	); err != nil {
		return
	}
	return
}

func (x *Transfer) CreateLogger(args app.CreateLoggerRequest) (err error) {
	if err = x.Client.Call(
		context.Background(),
		"CreateLogger",
		&args,
		&app.Empty{},
	); err != nil {
		return
	}
	return
}

func (x *Transfer) DeleteLogger(args app.DeleteLoggerRequest) (err error) {
	if err = x.Client.Call(
		context.Background(),
		"DeleteLogger",
		&args,
		&app.Empty{},
	); err != nil {
		return
	}
	return
}

func (x *Transfer) Publish(args app.PublishRequest) (err error) {
	if err = x.Client.Call(
		context.Background(),
		"Publish",
		&args,
		&app.Empty{},
	); err != nil {
		return
	}
	return
}
