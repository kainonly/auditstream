package controller

import (
	"context"
	"encoding/base64"
	"microtools-gossh/app/actions"
	"microtools-gossh/app/types"
	pb "microtools-gossh/router"
)

func (c *controller) Testing(ctx context.Context, param *pb.TestingParameter) (*pb.Response, error) {
	privateKey, err := base64.StdEncoding.DecodeString(param.PrivateKey)
	if err != nil {
		return c.response(err)
	}
	client, err := actions.Connect(types.SshOption{
		Host:       param.Host,
		Port:       param.Port,
		Username:   param.Username,
		Password:   param.Password,
		Key:        privateKey,
		PassPhrase: []byte(param.Passphrase),
	})
	if err != nil {
		return c.response(err)
	}
	defer client.Close()
	return c.response(nil)
}
