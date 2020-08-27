package controller

import (
	"context"
	"encoding/base64"
	"microtools-gossh/app/types"
	pb "microtools-gossh/router"
)

func (c *controller) Put(ctx context.Context, param *pb.PutParameter) (*pb.Response, error) {
	privateKey, err := base64.StdEncoding.DecodeString(param.PrivateKey)
	if err != nil {
		return c.response(err)
	}
	err = c.manager.Put(
		param.Identity,
		types.SshOption{
			Host:       param.Host,
			Port:       param.Port,
			Username:   param.Username,
			Password:   param.Password,
			Key:        privateKey,
			PassPhrase: []byte(param.Passphrase),
		},
	)
	if err != nil {
		return c.response(err)
	}
	return c.response(nil)
}
