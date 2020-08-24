package client

import (
	"encoding/base64"
	"errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"ssh-microservice/common"
	"sync"
)

type Client struct {
	options       map[string]*common.ConnectOption
	tunnels       map[string]*[]common.TunnelOption
	runtime       map[string]*ssh.Client
	localListener *syncMapListener
	localConn     *syncMapConn
	remoteConn    *syncMapConn
}

// Create ssh client service
func Create() *Client {
	var err error
	client := new(Client)
	client.options = make(map[string]*common.ConnectOption)
	client.tunnels = make(map[string]*[]common.TunnelOption)
	client.runtime = make(map[string]*ssh.Client)
	client.localListener = newSyncMapListener()
	client.localConn = newSyncMapConn()
	client.remoteConn = newSyncMapConn()
	var configs []common.ConfigOption
	configs, err = common.ListConfig()
	for _, opt := range configs {
		var key []byte
		key, err = base64.StdEncoding.DecodeString(opt.Key)
		if err != nil {
			log.Fatalln(err)
		}
		var passPhrase []byte
		passPhrase, err = base64.StdEncoding.DecodeString(opt.PassPhrase)
		if err != nil {
			log.Fatalln(err)
		}
		err = client.Put(opt.Identity, common.ConnectOption{
			Host:       opt.Host,
			Port:       opt.Port,
			Username:   opt.Username,
			Password:   opt.Password,
			Key:        key,
			PassPhrase: passPhrase,
		})
		if err != nil {
			log.Fatalln(err)
		}
		var tunnels []common.TunnelOption
		for _, val := range opt.Tunnels {
			tunnels = append(tunnels, common.TunnelOption{
				SrcIp:   val.SrcIp,
				SrcPort: val.SrcPort,
				DstIp:   val.DstIp,
				DstPort: val.DstPort,
			})
		}
		if len(tunnels) == 0 {
			continue
		}
		err = client.SetTunnels(opt.Identity, tunnels)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return client
}

// Generate Auth Method
func (c *Client) auth(option common.ConnectOption) (auth []ssh.AuthMethod, err error) {
	if len(option.Key) == 0 {
		// Password AuthMethod
		auth = []ssh.AuthMethod{
			ssh.Password(option.Password),
		}
	} else {
		// PrivateKey AuthMethod
		var signer ssh.Signer
		if len(option.PassPhrase) != 0 {
			// With Passphrase
			if signer, err = ssh.ParsePrivateKeyWithPassphrase(
				option.Key,
				option.PassPhrase,
			); err != nil {
				return
			}
		} else {
			// Without Passphrase
			if signer, err = ssh.ParsePrivateKey(
				option.Key,
			); err != nil {
				return
			}
		}
		auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	}
	return
}

// Ssh client connection
func (c *Client) connect(option common.ConnectOption) (client *ssh.Client, err error) {
	auth, err := c.auth(option)
	if err != nil {
		return
	}
	config := ssh.ClientConfig{
		User:            option.Username,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	addr := common.GetAddr(option.Host, uint(option.Port))
	client, err = ssh.Dial("tcp", addr, &config)
	return
}

func (c *Client) empty(identity string) error {
	if c.options[identity] == nil || c.runtime[identity] == nil {
		return errors.New("this identity does not exists")
	}
	return nil
}

// Test ssh client connection
func (c *Client) Testing(option common.ConnectOption) (*ssh.Client, error) {
	return c.connect(option)
}

// Add or modify the ssh client
func (c *Client) Put(identity string, option common.ConnectOption) (err error) {
	if c.options[identity] != nil && c.runtime[identity] != nil {
		if err = c.Delete(identity); err != nil {
			return
		}
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.options[identity] = &option
		c.runtime[identity], err = c.connect(option)
		if err != nil {
			return
		}
	}()
	wg.Wait()
	return common.SaveConfig(common.ConfigOption{
		Identity:   identity,
		Host:       option.Host,
		Port:       option.Port,
		Username:   option.Username,
		Password:   option.Password,
		Key:        base64.StdEncoding.EncodeToString(option.Key),
		PassPhrase: base64.StdEncoding.EncodeToString(option.PassPhrase),
		Tunnels:    nil,
	})
}

// Remotely execute commands via SSH
func (c *Client) Exec(identity string, cmd string) (output []byte, err error) {
	if err = c.empty(identity); err != nil {
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		session, err := c.runtime[identity].NewSession()
		if err != nil {
			return
		}
		defer session.Close()
		output, err = session.Output(cmd)
	}()
	wg.Wait()
	return
}

// Get ssh client information
func (c *Client) GetConnectOption(identity string) (option *common.ConnectOption, err error) {
	if err = c.empty(identity); err != nil {
		return
	}
	option = c.options[identity]
	return
}

func (c *Client) GetRuntime(identity string) (client *ssh.Client, err error) {
	if err = c.empty(identity); err != nil {
		return
	}
	client = c.runtime[identity]
	return
}

func (c *Client) GetTunnelOption(identity string) (option []common.TunnelOption, err error) {
	if err = c.empty(identity); err != nil {
		return
	}
	if c.tunnels[identity] != nil {
		option = *c.tunnels[identity]
	}
	return
}

func (c *Client) All() []string {
	var keys []string
	for key := range c.options {
		keys = append(keys, key)
	}
	return keys
}

// Delete ssh client
func (c *Client) Delete(identity string) (err error) {
	if err = c.empty(identity); err != nil {
		return
	}
	if c.tunnels[identity] != nil {
		c.closeTunnel(identity)
	}
	if c.runtime[identity] != nil {
		c.runtime[identity].Close()
	}
	delete(c.runtime, identity)
	delete(c.options, identity)
	return common.RemoveConfig(identity)
}

// Tunnel setting
func (c *Client) SetTunnels(identity string, options []common.TunnelOption) (err error) {
	if err = c.empty(identity); err != nil {
		return
	}
	if c.tunnels[identity] != nil {
		c.closeTunnel(identity)
	}
	c.tunnels[identity] = &options
	for _, tunnel := range options {
		go c.mutilTunnel(identity, tunnel)
	}
	return common.SaveConfig(common.ConfigOption{
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

// Close all running tunnels to which the identity belongs
func (c *Client) closeTunnel(identity string) {
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

// Multiple tunnel settings
func (c *Client) mutilTunnel(identity string, option common.TunnelOption) {
	localAddr := common.GetAddr(option.DstIp, uint(option.DstPort))
	remoteAddr := common.GetAddr(option.SrcIp, uint(option.SrcPort))
	localListener, err := net.Listen("tcp", localAddr)
	if err != nil {
		logrus.Error("<" + identity + ">:" + err.Error())
		return
	} else {
		c.localListener.Set(identity, localAddr, &localListener)
	}
	for {
		localConn, err := localListener.Accept()
		if err != nil {
			logrus.Error("<" + identity + ">:" + err.Error())
			return
		} else {
			c.localConn.Set(identity, localAddr, &localConn)
		}
		remoteConn, err := c.runtime[identity].Dial("tcp", remoteAddr)
		if err != nil {
			logrus.Error("remote <" + identity + ">:" + err.Error())
			return
		} else {
			c.remoteConn.Set(identity, localAddr, &remoteConn)
		}
		go c.forward(&localConn, &remoteConn)
	}
}

//  Tcp stream forwarding
func (c *Client) forward(local *net.Conn, remote *net.Conn) {
	defer (*local).Close()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		common.Copy(*local, *remote)
	}()
	go func() {
		defer wg.Done()
		if _, err := common.Copy(*remote, *local); err != nil {
			(*local).Close()
			(*remote).Close()
		}
		(*remote).Close()
	}()
	wg.Wait()
}
