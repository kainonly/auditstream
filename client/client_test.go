package client

import (
	"context"
	"fmt"
	"github.com/weplanx/transfer/bootstrap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"testing"
)

var x *Transfer

func TestMain(m *testing.M) {
	os.Chdir("../")
	v, err := bootstrap.SetValues()
	if err != nil {
		panic(err)
	}
	var host string
	var opts []grpc.DialOption
	if v.TLS.Cert != "" {
		creds, err := credentials.NewClientTLSFromFile(v.TLS.Cert, "")
		if err != nil {
			panic(err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
		host = "x.kainonly.com"
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		host = "127.0.0.1"
	}

	if x, err = New(fmt.Sprintf(`%s%s`, host, v.Address), opts...); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestTransfer_CreateLogger(t *testing.T) {
	defer x.Close()
	if err := x.CreateLogger(context.TODO(),
		"beta",
		"Transfer 测试",
	); err != nil {
		t.Error(err)
	}
}

func TestTransfer_Logger(t *testing.T) {
	defer x.Close()
	result, err := x.Logger(context.TODO())
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
}

func TestTransfer_DeleteLogger(t *testing.T) {
	defer x.Close()
	if err := x.DeleteLogger(context.TODO(),
		"312a2fc1-b758-454a-92ef-9c0bf6324fda",
	); err != nil {
		t.Error(err)
	}
}
