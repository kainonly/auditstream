package transfer_test

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/transfer"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
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
var mclient *mongo.Client

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
	if mclient, err = mongo.Connect(context.TODO(),
		options.Client().ApplyURI(os.Getenv("DATABASE")),
	); err != nil {
		log.Fatalln(err)
	}

	if client, err = transfer.New("test", mclient.Database("development"), js); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestTransfer_Set(t *testing.T) {
	err := client.Set(context.TODO(), transfer.Option{
		Key:         "system",
		Description: "测试",
		TTL:         3600,
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
	subject := fmt.Sprintf(`%s.logs.%s`, "test", "system")
	queue := fmt.Sprintf(`%s:logs:%s`, "test", "system")
	now := time.Now()
	go js.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		var payload transfer.Payload
		if err := sonic.Unmarshal(msg.Data, &payload); err != nil {
			t.Error(err)
		}
		t.Log(payload)
		assert.Equal(t, "0ff5483a-7ddc-44e0-b723-c3417988663f", payload.Metadata["uuid"])
		assert.Equal(t, map[string]interface{}{"msg": "hi"}, payload.Data)
		assert.Equal(t, now.UnixNano(), payload.Timestamp.UnixNano())
		wg.Done()
	})
	err := client.Publish(context.TODO(), "system", transfer.Payload{
		Metadata: map[string]interface{}{
			"uuid": "0ff5483a-7ddc-44e0-b723-c3417988663f",
		},
		Data: map[string]interface{}{
			"msg": "hi",
		},
		Timestamp: now,
	})
	assert.Nil(t, err)
	wg.Wait()
}

func TestTransfer_Remove(t *testing.T) {
	err := client.Remove("system")
	assert.Nil(t, err)
}
