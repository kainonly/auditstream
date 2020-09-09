package controller

import (
	"context"
	"elastic-transfer/app/types"
	pb "elastic-transfer/router"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var client pb.RouterClient

func TestMain(m *testing.M) {
	os.Chdir("../..")
	var err error
	if _, err := os.Stat("./config/autoload"); os.IsNotExist(err) {
		os.Mkdir("./config/autoload", os.ModeDir)
	}
	cfgByte, err := ioutil.ReadFile("./config/config.yml")
	if err != nil {
		log.Fatalln("Failed to read service configuration file", err)
	}
	config := types.Config{}
	err = yaml.Unmarshal(cfgByte, &config)
	if err != nil {
		log.Fatalln("Service configuration file parsing failed", err)
	}
	grpcConn, err := grpc.Dial(config.Listen, grpc.WithInsecure())
	if err != nil {
		logrus.Fatalln(err)
	}
	client = pb.NewRouterClient(grpcConn)
	os.Exit(m.Run())
}

func TestController_Put(t *testing.T) {
	response, err := client.Put(context.Background(), &pb.Information{
		Identity: "schedule",
		Index:    "schedule-logs",
		Validate: `{"type":"object"}`,
		Topic:    "sys.schedule",
		Key:      "",
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
	response, err = client.Put(context.Background(), &pb.Information{
		Identity: "mq-subscriber",
		Index:    "mq-subscriber-logs",
		Validate: `{"type":"object"}`,
		Topic:    "sys.subscriber",
		Key:      "",
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
}

func TestController_All(t *testing.T) {
	response, err := client.All(context.Background(), &pb.NoParameter{})
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
}

func TestController_Get(t *testing.T) {
	response, err := client.Get(context.Background(), &pb.GetParameter{
		Identity: "schedule",
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
}

func TestController_Lists(t *testing.T) {
	response, err := client.Lists(context.Background(), &pb.ListsParameter{
		Identity: []string{"schedule"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
}

func TestController_Push(t *testing.T) {
	response, err := client.Push(context.Background(), &pb.PushParameter{
		Identity: "schedule",
		Data:     []byte(`{"name":"kain"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
}

func BenchmarkController_Push(b *testing.B) {
	for i := 0; i < b.N; i++ {
		response, err := client.Push(context.Background(), &pb.PushParameter{
			Identity: "schedule",
			Data:     []byte(`{"name":"kain"}`),
		})
		if err != nil {
			b.Fatal(err)
		}
		if response.Error != 0 {
			b.Error(response.Msg)
		} else {
			b.Log(response.Msg)
		}
	}
}

func TestController_Delete(t *testing.T) {
	response, err := client.Delete(context.Background(), &pb.DeleteParameter{
		Identity: "schedule",
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
	response, err = client.Delete(context.Background(), &pb.DeleteParameter{
		Identity: "mq-subscriber",
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != 0 {
		t.Error(response.Msg)
	} else {
		t.Log(response.Msg)
	}
}
