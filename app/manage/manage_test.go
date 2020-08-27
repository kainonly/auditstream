package manage

import (
	"encoding/base64"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"microtools-gossh/app/types"
	"os"
	"testing"
)

var manager *ClientManager
var debug []*types.DebugOption

func TestMain(m *testing.M) {
	os.Chdir("../..")
	bytes, err := ioutil.ReadFile("./config/debug/config.yml")
	if err != nil {
		log.Fatalln(err)
	}
	err = yaml.Unmarshal(bytes, &debug)
	if err != nil {
		log.Fatalln(err)
	}
	bytes, err = ioutil.ReadFile("./config/debug/key-1.pem")
	if err != nil {
		log.Fatalln(err)
	}
	debug[0].PrivateKey = base64.StdEncoding.EncodeToString(bytes)
	manager, err = NewClientManager()
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(m.Run())
}

func TestClientManager_Put(t *testing.T) {
	option := debug[0]
	key, err := base64.StdEncoding.DecodeString(option.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	err = manager.Put("debug-1", types.SshOption{
		Host:       option.Host,
		Port:       option.Port,
		Username:   option.Username,
		Password:   option.Password,
		Key:        key,
		PassPhrase: []byte(option.Passphrase),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestClientManager_Exec(t *testing.T) {
	output, err := manager.Exec("debug-1", "date")
	if err != nil {
		t.Fatal(err)
	}
	if string(output) != "" {
		t.Log("output:", string(output))
	}
}

func TestClientManager_Tunnels(t *testing.T) {
	err := manager.Tunnels("debug-1", []types.TunnelOption{
		{
			SrcIp:   "127.0.0.1",
			SrcPort: 3306,
			DstIp:   "127.0.0.1",
			DstPort: 3306,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestClientManager_Delete(t *testing.T) {
	err := manager.Delete("debug-1")
	if err != nil {
		t.Fatal(err)
	}
}
