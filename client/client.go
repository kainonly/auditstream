package client

import "github.com/smallnest/rpcx/client"

type Transfer struct {
	Client client.XClient
}

func New(server string) (x *Transfer, err error) {
	x = new(Transfer)
	var discovery *client.Peer2PeerDiscovery
	if discovery, err = client.NewPeer2PeerDiscovery(server, ""); err != nil {
		return
	}
	x.Client = client.NewXClient("API",
		client.Failtry,
		client.RandomSelect,
		discovery,
		client.DefaultOption,
	)
	return
}
