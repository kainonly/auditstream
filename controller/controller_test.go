package controller

import (
	"context"
	"encoding/base64"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"ssh-microservice/common"
	pb "ssh-microservice/router"
	"testing"
)

var (
	conn  *grpc.ClientConn
	debug []*DebugOption
)

type DebugOption struct {
	Host       string `yaml:"host"`
	Port       uint32 `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	PrivateKey string `yaml:"private_key"`
	Passphrase string `yaml:"passphrase"`
}

func TestMain(m *testing.M) {
	os.Chdir("..")
	in, err := ioutil.ReadFile("./config/config.yml")
	if err != nil {
		log.Fatalln(err)
	}
	cfg := common.AppOption{}
	err = yaml.Unmarshal(in, &cfg)
	if err != nil {
		log.Fatalln(err)
	}
	in, err = ioutil.ReadFile("./config/debug.yml")
	if err != nil {
		log.Fatalln(err)
	}
	err = yaml.Unmarshal(in, &debug)
	if err != nil {
		log.Fatalln(err)
	}
	in, err = ioutil.ReadFile("./config/key-1.pem")
	if err != nil {
		log.Fatalln(err)
	}
	debug[0].PrivateKey = base64.StdEncoding.EncodeToString(in)
	conn, err = grpc.Dial(cfg.Listen, grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	in, err = ioutil.ReadFile("./config/key-2.pem")
	if err != nil {
		log.Fatalln(err)
	}
	debug[1].PrivateKey = base64.StdEncoding.EncodeToString(in)
	conn, err = grpc.Dial(cfg.Listen, grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(m.Run())
}

func TestConnect(t *testing.T) {
	defer conn.Close()
	client := pb.NewRouterClient(conn)
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
	}
	logrus.Info(response.Msg)
}

func TestPut(t *testing.T) {
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	response, err := client.Put(
		context.Background(),
		&pb.PutParameter{
			Identity:   "test",
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
	}
	logrus.Info(response.Msg)
}

func TestExec(t *testing.T) {
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	response, err := client.Exec(
		context.Background(),
		&pb.ExecParameter{
			Identity: "test",
			Bash:     "uptime",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	}
	logrus.Info(response.Data)
}

func BenchmarkExec(b *testing.B) {
	client := pb.NewRouterClient(conn)
	for i := 0; i < b.N; i++ {
		response, err := client.Exec(
			context.Background(),
			&pb.ExecParameter{
				Identity: "test",
				Bash:     "uptime",
			},
		)
		if err != nil {
			b.Fatal(err)
		}
		if response.Error != 0 {
			b.Error(response.Msg)
		}
		logrus.Info(response.Data)
	}
}

func TestPutOther(t *testing.T) {
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	response, err := client.Put(
		context.Background(),
		&pb.PutParameter{
			Identity:   "other",
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
	}
	logrus.Info(response.Msg)
}

func TestOtherExec(t *testing.T) {
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	response, err := client.Exec(
		context.Background(),
		&pb.ExecParameter{
			Identity: "other",
			Bash:     "uptime",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	}
	logrus.Info(response.Data)
}

func TestDelete(t *testing.T) {
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	response, err := client.Delete(
		context.Background(),
		&pb.DeleteParameter{
			Identity: "test",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	}
	logrus.Info(response.Msg)
}

func TestOtherDelete(t *testing.T) {
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	response, err := client.Delete(
		context.Background(),
		&pb.DeleteParameter{
			Identity: "other",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	}
	logrus.Info(response.Msg)
}

func TestGet(t *testing.T) {
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	response, err := client.Get(
		context.Background(),
		&pb.GetParameter{
			Identity: "test",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	}
	logrus.Info(response.Data)
}

func TestOtherGet(t *testing.T) {
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	response, err := client.Get(
		context.Background(),
		&pb.GetParameter{
			Identity: "other",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	}
	logrus.Info(response.Data)
}

func TestLists(t *testing.T) {
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	response, err := client.Lists(
		context.Background(),
		&pb.ListsParameter{
			Identity: []string{"test", "other"},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	}
	logrus.Info(response.Data)
}

func TestAll(t *testing.T) {
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	response, err := client.All(
		context.Background(),
		&pb.NoParameter{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	}
	logrus.Info(response.Data)
}

func TestTunnels(t *testing.T) {
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	response, err := client.Tunnels(
		context.Background(),
		&pb.TunnelsParameter{
			Identity: "test",
			Tunnels: []*pb.TunnelOption{
				&pb.TunnelOption{
					SrcIp:   "127.0.0.1",
					SrcPort: 3306,
					DstIp:   "127.0.0.1",
					DstPort: 3306,
				},
				&pb.TunnelOption{
					SrcIp:   "127.0.0.1",
					SrcPort: 9200,
					DstIp:   "127.0.0.1",
					DstPort: 9200,
				},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	}
	logrus.Info(response.Msg)
}

func TestOtherTunnels(t *testing.T) {
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	response, err := client.Tunnels(
		context.Background(),
		&pb.TunnelsParameter{
			Identity: "other",
			Tunnels: []*pb.TunnelOption{
				&pb.TunnelOption{
					SrcIp:   "127.0.0.1",
					SrcPort: 80,
					DstIp:   "127.0.0.1",
					DstPort: 8080,
				},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	}
	logrus.Info(response.Msg)
}
