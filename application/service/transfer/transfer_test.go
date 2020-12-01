package transfer

import (
	"elastic-transfer/application/service/schema"
	"elastic-transfer/config"
	"elastic-transfer/config/options"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var es *Elastic

func TestMain(m *testing.M) {
	os.Chdir("../../../")
	var err error
	var cfg *config.Config
	if _, err = os.Stat("./config/autoload"); os.IsNotExist(err) {
		os.Mkdir("./config/autoload", os.ModeDir)
	}
	if _, err = os.Stat("./config/config.yml"); os.IsNotExist(err) {
		log.Fatalln(err)
	}
	var bs []byte
	if bs, err = ioutil.ReadFile("./config/config.yml"); err != nil {
		return
	}
	if err = yaml.Unmarshal(bs, &cfg); err != nil {
		return
	}
	if es, err = New(&Dependency{
		Config: cfg,
		Schema: schema.New("./config/autoload/"),
	}); err != nil {
		log.Fatalln(err)
	}
	os.Exit(m.Run())
}

func TestElasticManager_Put(t *testing.T) {
	err := es.Put(options.PipeOption{
		Identity: "debug",
		Index:    "debug-log",
		Validate: `{"type":"object","properties":{"name":{"type":"string"}}}`,
		Topic:    "sys.debug",
		Key:      "",
	})
	if err != nil {
		t.Error(err)
	}
}

func TestElasticManager_Delete(t *testing.T) {
	err := es.Delete("debug")
	if err != nil {
		t.Error(err)
	}
}
