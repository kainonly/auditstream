package controller

import (
	"context"
	pb "microtools-gossh/router"
)

func (c *controller) Get(ctx context.Context, req *pb.GetParameter) (response *pb.GetResponse, err error) {
	sshOption, err := c.manager.GetSshOption(req.Identity)
	if err != nil {
		return c.getErrorResponse(err)
	}
	client, err := c.manager.GetRuntime(req.Identity)
	if err != nil {
		return c.getErrorResponse(err)
	}
	tunnelOption, err := c.manager.GetTunnelOption(req.Identity)
	if err != nil {
		return c.getErrorResponse(err)
	}
	resultTunnelOption := make([]*pb.TunnelOption, len(tunnelOption))
	for index, option := range tunnelOption {
		resultTunnelOption[index] = &pb.TunnelOption{
			SrcIp:   option.SrcIp,
			SrcPort: option.SrcPort,
			DstIp:   option.DstIp,
			DstPort: option.DstPort,
		}
	}
	return c.getSuccessResponse(&pb.Information{
		Identity:  req.Identity,
		Host:      sshOption.Host,
		Port:      sshOption.Port,
		Username:  sshOption.Username,
		Connected: string(client.ClientVersion()),
		Tunnels:   resultTunnelOption,
	})
}

func (c *controller) getErrorResponse(err error) (*pb.GetResponse, error) {
	return &pb.GetResponse{
		Error: 1,
		Msg:   err.Error(),
	}, nil
}

func (c *controller) getSuccessResponse(data *pb.Information) (*pb.GetResponse, error) {
	return &pb.GetResponse{
		Error: 0,
		Msg:   "ok",
		Data:  data,
	}, nil
}
