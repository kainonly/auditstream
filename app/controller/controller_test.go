package controller

import (
	"context"
	"encoding/base64"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"microtools-gossh/app/types"
	pb "microtools-gossh/router"
	"os"
	"testing"
	"time"
)

var debug []*types.DebugOption
var client pb.RouterClient

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
	bytes, err = ioutil.ReadFile("./config/debug/key-2.pem")
	if err != nil {
		log.Fatalln(err)
	}
	debug[1].PrivateKey = base64.StdEncoding.EncodeToString(bytes)
	if _, err := os.Stat("./config/autoload"); os.IsNotExist(err) {
		os.Mkdir("./config/autoload", os.ModeDir)
	}
	if _, err := os.Stat("./config/config.yml"); os.IsNotExist(err) {
		logrus.Fatalln("The service configuration file does not exist")
	}
	cfgByte, err := ioutil.ReadFile("./config/config.yml")
	if err != nil {
		logrus.Fatalln("Failed to read service configuration file", err)
	}
	config := types.Config{}
	err = yaml.Unmarshal(cfgByte, &config)
	if err != nil {
		logrus.Fatalln("Service configuration file parsing failed", err)
	}
	grpcConn, err := grpc.Dial(config.Listen, grpc.WithInsecure())
	if err != nil {
		logrus.Fatalln(err)
	}
	client = pb.NewRouterClient(grpcConn)
	os.Exit(m.Run())
}

func TestController_Testing(t *testing.T) {
	response, err := client.Testing(
		context.Background(),
		&pb.TestingParameter{
			Host:       debug[0].Host,
			Port:       debug[0].Port,
			Username:   debug[0].Username,
			Password:   debug[0].Password,
			PrivateKey: debug[0].PrivateKey,
			Passphrase: debug[0].Passphrase,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
}

func TestController_Put(t *testing.T) {
	response, err := client.Put(
		context.Background(),
		&pb.PutParameter{
			Identity:   "debug-1",
			Host:       debug[0].Host,
			Port:       debug[0].Port,
			Username:   debug[0].Username,
			Password:   debug[0].Password,
			PrivateKey: debug[0].PrivateKey,
			Passphrase: debug[0].Passphrase,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
	response, err = client.Put(
		context.Background(),
		&pb.PutParameter{
			Identity:   "debug-2",
			Host:       debug[1].Host,
			Port:       debug[1].Port,
			Username:   debug[1].Username,
			Password:   debug[1].Password,
			PrivateKey: debug[1].PrivateKey,
			Passphrase: debug[1].Passphrase,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
}

func TestController_Exec(t *testing.T) {
	response, err := client.Exec(
		context.Background(),
		&pb.ExecParameter{
			Identity: "debug-1",
			Bash:     "uptime",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Data)
	}
	response, err = client.Exec(
		context.Background(),
		&pb.ExecParameter{
			Identity: "debug-2",
			Bash:     "uptime",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Data)
	}
}

func TestController_Get(t *testing.T) {
	response, err := client.Get(
		context.Background(),
		&pb.GetParameter{
			Identity: "debug-1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Data)
	}
}

func TestController_All(t *testing.T) {
	response, err := client.All(
		context.Background(),
		&pb.NoParameter{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Data)
	}
}

func TestController_Lists(t *testing.T) {
	response, err := client.Lists(
		context.Background(),
		&pb.ListsParameter{
			Identity: []string{"debug-1", "debug-2"},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Data)
	}
}

func TestController_Tunnels(t *testing.T) {
	time.Sleep(time.Second)
	response, err := client.Tunnels(
		context.Background(),
		&pb.TunnelsParameter{
			Identity: "debug-1",
			Tunnels: []*pb.TunnelOption{
				{
					SrcIp:   "127.0.0.1",
					SrcPort: 9200,
					DstIp:   "127.0.0.1",
					DstPort: 9200,
				},
				{
					SrcIp:   "127.0.0.1",
					SrcPort: 5601,
					DstIp:   "127.0.0.1",
					DstPort: 5601,
				},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
	response, err = client.Tunnels(
		context.Background(),
		&pb.TunnelsParameter{
			Identity: "debug-2",
			Tunnels: []*pb.TunnelOption{
				{
					SrcIp:   "127.0.0.1",
					SrcPort: 80,
					DstIp:   "127.0.0.1",
					DstPort: 80,
				},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
}

func BenchmarkController_Exec(b *testing.B) {
	b.RunParallel(func(testPB *testing.PB) {
		for testPB.Next() {
			_, err := client.Exec(
				context.Background(),
				&pb.ExecParameter{
					Identity: "debug-1",
					Bash:     "uptime",
				},
			)
			if err != nil {
				b.Fatal(err)
			}
			_, err = client.Exec(
				context.Background(),
				&pb.ExecParameter{
					Identity: "debug-2",
					Bash:     "uptime",
				},
			)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.ReportAllocs()
}

func TestController_Delete(t *testing.T) {
	response, err := client.Delete(
		context.Background(),
		&pb.DeleteParameter{
			Identity: "debug-1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
	response, err = client.Delete(
		context.Background(),
		&pb.DeleteParameter{
			Identity: "debug-2",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
}
