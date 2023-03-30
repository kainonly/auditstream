package transfer_test

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/weplanx/transfer"
	"os"
	"sync"
	"testing"
	"time"
)

import (
	"context"
)

var client *transfer.Transfer
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
			panic("nkey is invalid")
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
	if client, err = transfer.New(
		transfer.SetNamespace("example"),
		transfer.SetJetStream(js),
	); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestTransfer_Set(t *testing.T) {
	err := client.Set(context.TODO(), transfer.LogOption{
		Key:         "system",
		Description: "system beta",
	})
	assert.Nil(t, err)
}

func TestTransfer_Get(t *testing.T) {
	_, err := client.Get("not_exists")
	assert.Error(t, err)
	result, err := client.Get("system")
	assert.Nil(t, err)
	t.Log(result)
}

func TestTransfer_Publish(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	subjectName := fmt.Sprintf(`%s.logs.%s`, "beta", "system")
	queueName := fmt.Sprintf(`%s:logs:%s`, "beta", "system")
	now := time.Now()
	go js.QueueSubscribe(subjectName, queueName, func(msg *nats.Msg) {
		var payload transfer.Payload
		if err := msgpack.Unmarshal(msg.Data, &payload); err != nil {
			t.Error(err)
		}
		t.Log(payload)
		assert.Equal(t, "0ff5483a-7ddc-44e0-b723-c3417988663f", payload.Data["uuid"])
		assert.Equal(t, map[string]interface{}{"msg": "hi"}, payload.Data)
		assert.Equal(t, now.UnixNano(), payload.Timestamp.UnixNano())
		wg.Done()
	})
	err := client.Publish(context.TODO(), "system", transfer.Payload{
		Data: map[string]interface{}{
			"uuid": "0ff5483a-7ddc-44e0-b723-c3417988663f",
			"msg":  "hi",
		},
		Timestamp: now,
	})
	assert.NoError(t, err)
	wg.Wait()
}

//func TestTransfer_ManualPublish(t *testing.T) {
//	now := time.Now()
//	err := client.Publish(context.TODO(), "system", transfer.Payload{
//		Timestamp: now,
//		Data: map[string]interface{}{
//			"metadata": map[string]interface{}{
//				"user_id": "640e7c2c7d8a24d6f831e9bf",
//			},
//			"msg": "123456",
//		},
//		Format: map[string]interface{}{
//			"metadata.user_id": "oid",
//		},
//	})
//	assert.NoError(t, err)
//}

func TestTransfer_Remove(t *testing.T) {
	err := client.Remove("system")
	assert.Nil(t, err)
}
