package controller

import (
	"context"
	"microtools-gossh/app/types"
	pb "microtools-gossh/router"
)

func (c *controller) Tunnels(ctx context.Context, param *pb.TunnelsParameter) (*pb.Response, error) {
	tunnelOptions := make([]types.TunnelOption, len(param.Tunnels))
	for index, option := range param.Tunnels {
		tunnelOptions[index] = types.TunnelOption{
			SrcIp:   option.SrcIp,
			SrcPort: option.SrcPort,
			DstIp:   option.DstIp,
			DstPort: option.DstPort,
		}
	}
	err := c.manager.Tunnels(param.Identity, tunnelOptions)
	if err != nil {
		return c.response(err)
	}
	return c.response(nil)
}
