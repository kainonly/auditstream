package client

import (
	"context"
	"fmt"
	"github.com/weplanx/transfer/bootstrap"
	"google.golang.org/grpc"
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
	if x, err = New(
		fmt.Sprintf(`127.0.0.1%s`, v.Address),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	); err != nil {
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
		"f4d165c2-1155-49e6-afb2-b2992d7c6bd3",
	); err != nil {
		t.Error(err)
	}
}
