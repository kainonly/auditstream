package transfer

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
	"os"
	"sync"
	"testing"
	"time"
)

var x *Transfer
var js nats.JetStreamContext

func TestMain(m *testing.M) {
	var err error
	token := os.Getenv("TOKEN")
	var auth nats.Option
	if token != "" {
		auth = nats.Token(token)
	} else {
		var kp nkeys.KeyPair
		if kp, err = nkeys.FromSeed([]byte(os.Getenv("NKEY"))); err != nil {
			return
		}
		defer kp.Wipe()
		var pub string
		if pub, err = kp.PublicKey(); err != nil {
			return
		}
		if !nkeys.IsValidPublicUserKey(pub) {
			panic("nkey 验证失败")
		}
		auth = nats.Nkey(pub, func(nonce []byte) ([]byte, error) {
			sig, _ := kp.Sign(nonce)
			return sig, nil
		})
	}
	var nc *nats.Conn
	if nc, err = nats.Connect(
		os.Getenv("HOSTS"),
		nats.MaxReconnects(5),
		nats.ReconnectWait(2*time.Second),
		nats.ReconnectJitter(500*time.Millisecond, 2*time.Second),
		auth,
	); err != nil {
		return
	}
	if js, err = nc.JetStream(nats.PublishAsyncMaxPending(256)); err != nil {
		panic(err)
	}
	if x, err = New("alpha", js); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

var key = "e2066c57-5669-d2d8-243e-ba19a6c18c45"

func TestTransfer_Set(t *testing.T) {
	if err := x.Set(key, Option{
		Topic:       "system",
		Description: "测试",
	}); err != nil {
		t.Error(err)
	}
}

func TestTransfer_Get(t *testing.T) {
	result, err := x.Get(key)
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
}

func TestTransfer_Publish(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	subject := fmt.Sprintf(`%s.logs.%s`, "alpha", "system")
	queue := fmt.Sprintf(`%s:logs:%s`, "alpha", "system")
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
	payload, err := NewPayload(CLSDto{
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

func TestTransfer_Remove(t *testing.T) {
	if err := x.Remove(key); err != nil {
		t.Error(err)
	}
}
