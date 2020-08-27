package actions

import (
	"encoding/base64"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"microtools-gossh/app/types"
	"os"
	"testing"
)

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
	os.Exit(m.Run())
}
