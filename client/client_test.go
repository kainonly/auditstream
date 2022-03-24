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

func TestTransfer_Create(t *testing.T) {
	if err := x.Create(context.TODO(),
		key,
		"system",
		"Transfer 新增",
	); err != nil {
		t.Error(err)
	}
}

func TestTransfer_Get(t *testing.T) {
	result, err := x.Get(context.TODO())
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
}

func TestTransfer_Update(t *testing.T) {
	if err := x.Update(context.TODO(), key,
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
	subject := fmt.Sprintf(`%s.logs.%s`, v.Namespace, "system")
	queue := fmt.Sprintf(`%s:logs:%s`, v.Namespace, "system")
	now := time.Now()
	go js.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		var data CLSDto
		if err := msgpack.Unmarshal(msg.Data, &data); err != nil {
			t.Error(err)
		}
		t.Log(data)
		assert.Equal(t, "0ff5483a-7ddc-44e0-b723-c3417988663f", data.TopicId)
		assert.Equal(t, map[string]string{"msg": "hi"}, data.Record)
		assert.Equal(t, now.Unix(), data.Time.Unix())
		wg.Done()
	})
	payload, err := NewPayload[CLSDto](CLSDto{
		TopicId: "0ff5483a-7ddc-44e0-b723-c3417988663f",
		Record: map[string]string{
			"msg": "hi",
		},
		Time: now,
	})
	if err != nil {
		t.Error(err)
	}
	if err := x.Publish(context.TODO(), "system", payload); err != nil {
		t.Error(err)
	}
	wg.Wait()
}

func TestTransfer_Delete(t *testing.T) {
	if err := x.Delete(context.TODO(), key); err != nil {
		t.Error(err)
	}
}
