package client

import (
	"fmt"
	"github.com/weplanx/transfer/app"
	"github.com/weplanx/transfer/bootstrap"
	"os"
	"testing"
)

var x *Transfer

func TestMain(m *testing.M) {
	os.Chdir("../")
	values, err := bootstrap.SetValues()
	if err != nil {
		panic(err)
	}
	if x, err = New(
		fmt.Sprintf(`transfer.kainonly.com%s`, values.Address),
		values.TLS,
	); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestTransfer_Logger(t *testing.T) {
	defer x.Close()
	result, err := x.Logger()
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
}

func TestTransfer_CreateLogger(t *testing.T) {
	defer x.Close()
	if err := x.CreateLogger(app.CreateLoggerRequest{
		Topic:       "beta",
		Description: "测试",
	}); err != nil {
		t.Error(err)
	}
}

func TestTransfer_DeleteLogger(t *testing.T) {
	defer x.Close()
	//if err:=x.DeleteLogger(app.DeleteLoggerRequest{Id: t})
}
