package schema

import (
	"microtools-gossh/app/types"
	"os"
	"testing"
)

var schema *Schema

func TestMain(m *testing.M) {
	os.Chdir("../..")
	schema = New()
	os.Exit(m.Run())
}

func TestSchema_Update(t *testing.T) {
	err := schema.Update(types.ClientOption{
		Identity:   "test",
		Host:       "127.0.0.1",
		Port:       22,
		Username:   "root",
		Password:   "123456",
		Key:        "",
		PassPhrase: "",
		Tunnels:    nil,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSchema_Lists(t *testing.T) {
	_, err := schema.Lists()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSchema_Delete(t *testing.T) {
	err := schema.Delete("test")
	if err != nil {
		t.Fatal(err)
	}
}
