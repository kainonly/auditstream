package api

import (
	"context"
	"github.com/smallnest/rpcx/client"
	"os"
	"testing"
)

var xclient client.XClient

func TestMain(m *testing.M) {
	d, err := client.NewPeer2PeerDiscovery("tcp@127.0.0.1:8972", "")
	if err != nil {
		panic(err)
	}
	xclient = client.NewXClient("API", client.Failtry, client.RandomSelect, d, client.DefaultOption)
	os.Exit(m.Run())
}

func TestAPI_Send(t *testing.T) {
	args := &SendArgs{
		Code: 123,
	}

	reply := &Reply{}
	err := xclient.Call(context.Background(), "Send", args, reply)
	if err != nil {
		t.Error(err)
	}
	t.Log(reply)
}
