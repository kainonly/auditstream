package manage

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"transfer-microservice/app/types"
)

var manager *ElasticManager

func TestMain(m *testing.M) {
	os.Chdir("../..")
	var err error
	cfgByte, err := ioutil.ReadFile("./config/config.yml")
	if err != nil {
		log.Fatalln("Failed to read service configuration file", err)
	}
	config := types.Config{}
	err = yaml.Unmarshal(cfgByte, &config)
	if err != nil {
		log.Fatalln("Service configuration file parsing failed", err)
	}
	manager, err = NewElasticManager(config.Elastic)
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(m.Run())
}

func TestElasticManager_Index(t *testing.T) {
	err := manager.Index("test", []byte(`{"name":"kain"}`))
	if err != nil {
		t.Fatal(err)
	}
}
