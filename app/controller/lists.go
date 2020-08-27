package controller

import (
	"context"
	pb "microtools-gossh/router"
)

func (c *controller) Lists(ctx context.Context, param *pb.ListsParameter) (*pb.ListsResponse, error) {
	lists := make([]*pb.Information, len(param.Identity))
	for index, identity := range param.Identity {
		sshOption, err := c.manager.GetSshOption(identity)
		if err != nil {
			return c.listsErrorResponse(err)
		}
		client, err := c.manager.GetRuntime(identity)
		if err != nil {
			return c.listsErrorResponse(err)
		}
		tunnelOption, err := c.manager.GetTunnelOption(identity)
		if err != nil {
			return c.listsErrorResponse(err)
		}
		resultTunnelOption := make([]*pb.TunnelOption, len(tunnelOption))
		for tIndex, option := range tunnelOption {
			resultTunnelOption[tIndex] = &pb.TunnelOption{
				SrcIp:   option.SrcIp,
				SrcPort: option.SrcPort,
				DstIp:   option.DstIp,
				DstPort: option.DstPort,
			}
		}
		lists[index] = &pb.Information{
			Identity:  identity,
			Host:      sshOption.Host,
			Port:      sshOption.Port,
			Username:  sshOption.Username,
			Connected: string(client.ClientVersion()),
			Tunnels:   resultTunnelOption,
		}
	}
	return c.listsSuccessResponse(lists)
}

func (c *controller) listsErrorResponse(err error) (*pb.ListsResponse, error) {
	return &pb.ListsResponse{
		Error: 1,
		Msg:   err.Error(),
	}, nil
}

func (c *controller) listsSuccessResponse(data []*pb.Information) (*pb.ListsResponse, error) {
	return &pb.ListsResponse{
		Error: 0,
		Msg:   "ok",
		Data:  data,
	}, nil
}
