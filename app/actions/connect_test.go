package actions

import (
	"encoding/base64"
	"microtools-gossh/app/types"
	"testing"
)

func TestConnect(t *testing.T) {
	option := debug[0]
	key, err := base64.StdEncoding.DecodeString(option.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	client, err := Connect(types.SshOption{
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
	if string(client.ClientVersion()) != "" {
		t.Logf("Successfully connected via go-ssh")
	}
}
