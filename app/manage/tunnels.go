package manage

import (
	"encoding/base64"
	"io"
	"microtools-gossh/app/types"
	"microtools-gossh/app/utils"
	"net"
	"sync"
)

func (c *ClientManager) Tunnels(identity string, options []types.TunnelOption) (err error) {
	if err = c.empty(identity); err != nil {
		return
	}
	if c.tunnels[identity] != nil {
		c.closeTunnel(identity)
	}
	c.tunnels[identity] = &options
	for _, tunnel := range options {
		go c.setTunnel(identity, tunnel)
	}
	return c.schema.Update(types.ClientOption{
		Identity:   identity,
		Host:       c.options[identity].Host,
		Port:       c.options[identity].Port,
		Username:   c.options[identity].Username,
		Password:   c.options[identity].Password,
		Key:        base64.StdEncoding.EncodeToString(c.options[identity].Key),
		PassPhrase: base64.StdEncoding.EncodeToString(c.options[identity].PassPhrase),
		Tunnels:    options,
	})
}

// Multiple tunnel settings
func (c *ClientManager) setTunnel(identity string, option types.TunnelOption) {
	localAddr := utils.GetAddr(option.DstIp, uint(option.DstPort))
	remoteAddr := utils.GetAddr(option.SrcIp, uint(option.SrcPort))
	localListener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return
	}
	c.localListener.Set(identity, localAddr, &localListener)
	for {
		localConn, err := localListener.Accept()
		if err != nil {
			return
		}
		c.localConn.Set(identity, localAddr, &localConn)
		remoteConn, err := c.runtime[identity].Dial("tcp", remoteAddr)
		if err != nil {
			return
		}
		c.remoteConn.Set(identity, localAddr, &remoteConn)
		go c.forward(&localConn, &remoteConn)
	}
}

//  Tcp stream forwarding
func (c *ClientManager) forward(local *net.Conn, remote *net.Conn) {
	defer (*local).Close()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(*local, *remote)
	}()
	go func() {
		defer wg.Done()
		if _, err := io.Copy(*remote, *local); err != nil {
			(*local).Close()
			(*remote).Close()
		}
		(*remote).Close()
	}()
	wg.Wait()
}

// Close all running tunnels to which the identity belongs
func (c *ClientManager) closeTunnel(identity string) {
	for _, conn := range c.remoteConn.Map[identity] {
		(*conn).Close()
	}
	c.remoteConn.Clear(identity)
	for _, conn := range c.localConn.Map[identity] {
		(*conn).Close()
	}
	c.localConn.Clear(identity)
	for _, listener := range c.localListener.Map[identity] {
		_ = (*listener).Close()
	}
	c.localListener.Clear(identity)
}
