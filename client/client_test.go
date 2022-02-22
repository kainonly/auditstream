package client

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/weplanx/transfer/bootstrap"
	"github.com/weplanx/transfer/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"sync"
	"testing"
	"time"
)

var v *common.Values
var x *Transfer

func TestMain(m *testing.M) {
	os.Chdir("../")
	var err error
	if v, err = bootstrap.SetValues(); err != nil {
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

var key = "e2066c57-5669-d2d8-243e-ba19a6c18c45"

func TestTransfer_CreateLogger(t *testing.T) {
	if err := x.CreateLogger(context.TODO(),
		key,
		"system",
		"Transfer 新增",
	); err != nil {
		t.Error(err)
	}
}

func TestTransfer_GetLoggers(t *testing.T) {
	result, err := x.GetLoggers(context.TODO())
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
	key = result[0].Key
}

func TestTransfer_UpdateLogger(t *testing.T) {
	if err := x.UpdateLogger(context.TODO(), key,
		"Transfer 工具",
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

func TestTransfer_Publish(t *testing.T) {
	nc, err := bootstrap.UseNats(v)
	if err != nil {
		t.Error(err)
	}
	js, err := bootstrap.UseJetStream(nc)
	if err != nil {
		t.Error(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	subject := fmt.Sprintf(`logs.%s.%s`, v.Namespace, "system")
	go js.Subscribe(subject, func(msg *nats.Msg) {
		var v map[string]interface{}
		if err := msgpack.Unmarshal(msg.Data, &v); err != nil {
			t.Error(err)
		}
		t.Log(v)
		assert.Equal(t, "hello", v["msg"])
		wg.Done()
	})
	if err := x.Publish(context.TODO(), "system", map[string]interface{}{
		"msg":  "hello",
		"time": time.Now(),
	}); err != nil {
		t.Error(err)
	}
	wg.Wait()
}

func TestTransfer_DeleteLogger(t *testing.T) {
	if err := x.DeleteLogger(context.TODO(), key); err != nil {
		t.Error(err)
	}
}
