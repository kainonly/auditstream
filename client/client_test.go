package client

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
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
	if err := x.CreateLogger(context.TODO(),
		"beta",
		"Transfer 新增",
	); err != nil {
		t.Error(err)
	}
}

var key string

func TestTransfer_GetLoggers(t *testing.T) {
	result, err := x.GetLoggers(context.TODO())
	if err != nil {
		t.Error(err)
	}
	assert.Len(t, result, 1)
	t.Log(result)
	key = result[0].Key
}

func TestTransfer_UpdateLogger(t *testing.T) {
	if err := x.UpdateLogger(context.TODO(), key,
		"beta1",
		"Transfer 修改",
	); err != nil {
		t.Error(err)
	}
}

func TestTransfer_Info(t *testing.T) {
	result, err := x.Info(context.TODO(), key)
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
}

func TestTransfer_DeleteLogger(t *testing.T) {
	if err := x.DeleteLogger(context.TODO(), key); err != nil {
		t.Error(err)
	}
}
