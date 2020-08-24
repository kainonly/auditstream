package controller

import (
	"context"
	"encoding/base64"
	"ssh-microservice/client"
	"ssh-microservice/common"
	pb "ssh-microservice/router"
)

type controller struct {
	pb.UnimplementedRouterServer
	client client.Client
}

func New(client *client.Client) *controller {
	c := new(controller)
	c.client = *client
	return c
}

func (c *controller) Testing(ctx context.Context, req *pb.TestingParameter) (*pb.Response, error) {
	privateKey, err := base64.StdEncoding.DecodeString(req.PrivateKey)
	if err != nil {
		return &pb.Response{
			Error: 1,
			Msg:   err.Error(),
		}, nil
	}
	cli, err := c.client.Testing(common.ConnectOption{
		Host:       req.Host,
		Port:       req.Port,
		Username:   req.Username,
		Password:   req.Password,
		Key:        privateKey,
		PassPhrase: []byte(req.Passphrase),
	})
	if err != nil {
		return &pb.Response{
			Error: 1,
			Msg:   err.Error(),
		}, nil
	}
	defer cli.Close()
	return &pb.Response{
		Error: 0,
		Msg:   "ok",
	}, nil
}

func (c *controller) Put(ctx context.Context, req *pb.PutParameter) (*pb.Response, error) {
	privateKey, err := base64.StdEncoding.DecodeString(req.PrivateKey)
	if err != nil {
		return &pb.Response{
			Error: 1,
			Msg:   err.Error(),
		}, nil
	}
	err = c.client.Put(
		req.Identity,
		common.ConnectOption{
			Host:       req.Host,
			Port:       req.Port,
			Username:   req.Username,
			Password:   req.Password,
			Key:        privateKey,
			PassPhrase: []byte(req.Passphrase),
		},
	)
	if err != nil {
		return &pb.Response{
			Error: 1,
			Msg:   err.Error(),
		}, nil
	}
	return &pb.Response{
		Error: 0,
		Msg:   "ok",
	}, nil
}

func (c *controller) Exec(ctx context.Context, params *pb.ExecParameter) (*pb.ExecResponse, error) {
	output, err := c.client.Exec(params.Identity, params.Bash)
	if err != nil {
		return &pb.ExecResponse{
			Error: 1,
			Msg:   err.Error(),
		}, nil
	}
	return &pb.ExecResponse{
		Error: 0,
		Data:  string(output),
	}, nil
}

func (c *controller) Delete(ctx context.Context, params *pb.DeleteParameter) (*pb.Response, error) {
	err := c.client.Delete(params.Identity)
	if err != nil {
		return &pb.Response{
			Error: 1,
			Msg:   err.Error(),
		}, nil
	}
	return &pb.Response{
		Error: 0,
		Msg:   "ok",
	}, nil
}

func (c *controller) Get(ctx context.Context, req *pb.GetParameter) (response *pb.GetResponse, err error) {
	connect, err := c.client.GetConnectOption(req.Identity)
	if err != nil {
		return &pb.GetResponse{
			Error: 1,
			Msg:   err.Error(),
		}, nil
	}
	cli, err := c.client.GetRuntime(req.Identity)
	if err != nil {
		return &pb.GetResponse{
			Error: 1,
			Msg:   err.Error(),
		}, nil
	}
	tunnel, err := c.client.GetTunnelOption(req.Identity)
	if err != nil {
		return &pb.GetResponse{
			Error: 1,
			Msg:   err.Error(),
		}, nil
	}
	var pbTunnelOption []*pb.TunnelOption
	for _, option := range tunnel {
		pbTunnelOption = append(pbTunnelOption, &pb.TunnelOption{
			SrcIp:   option.SrcIp,
			SrcPort: option.SrcPort,
			DstIp:   option.DstIp,
			DstPort: option.DstPort,
		})
	}
	return &pb.GetResponse{
		Error: 0,
		Data: &pb.Information{
			Identity:  req.Identity,
			Host:      connect.Host,
			Port:      connect.Port,
			Username:  connect.Username,
			Connected: string(cli.ClientVersion()),
			Tunnels:   pbTunnelOption,
		},
	}, nil
}

func (c *controller) Lists(ctx context.Context, req *pb.ListsParameter) (*pb.ListsResponse, error) {
	var lists []*pb.Information
	for _, identity := range req.Identity {
		connect, err := c.client.GetConnectOption(identity)
		if err != nil {
			return &pb.ListsResponse{
				Error: 1,
				Msg:   err.Error(),
			}, nil
		}
		cli, err := c.client.GetRuntime(identity)
		if err != nil {
			return &pb.ListsResponse{
				Error: 1,
				Msg:   err.Error(),
			}, nil
		}
		tunnel, err := c.client.GetTunnelOption(identity)
		if err != nil {
			return &pb.ListsResponse{
				Error: 1,
				Msg:   err.Error(),
			}, nil
		}
		var pbTunnelOption []*pb.TunnelOption
		for _, option := range tunnel {
			pbTunnelOption = append(pbTunnelOption, &pb.TunnelOption{
				SrcIp:   option.SrcIp,
				SrcPort: option.SrcPort,
				DstIp:   option.DstIp,
				DstPort: option.DstPort,
			})
		}
		lists = append(lists, &pb.Information{
			Identity:  identity,
			Host:      connect.Host,
			Port:      connect.Port,
			Username:  connect.Username,
			Connected: string(cli.ClientVersion()),
			Tunnels:   pbTunnelOption,
		})
	}
	return &pb.ListsResponse{
		Error: 0,
		Data:  lists,
	}, nil
}

func (c *controller) All(ctx context.Context, req *pb.NoParameter) (*pb.AllResponse, error) {
	return &pb.AllResponse{
		Error: 0,
		Data:  c.client.All(),
	}, nil
}

func (c *controller) Tunnels(ctx context.Context, req *pb.TunnelsParameter) (*pb.Response, error) {
	var tunnels []common.TunnelOption
	for _, value := range req.Tunnels {
		tunnels = append(tunnels, common.TunnelOption{
			SrcIp:   value.SrcIp,
			SrcPort: value.SrcPort,
			DstIp:   value.DstIp,
			DstPort: value.DstPort,
		})
	}
	err := c.client.SetTunnels(req.Identity, tunnels)
	if err != nil {
		return &pb.Response{
			Error: 1,
			Msg:   err.Error(),
		}, nil
	}
	return &pb.Response{
		Error: 0,
		Msg:   "ok",
	}, nil
}
