package manage

import (
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/ssh"
	"microtools-gossh/app/schema"
	"microtools-gossh/app/types"
	"microtools-gossh/app/utils"
)

type ClientManager struct {
	options       map[string]*types.SshOption
	tunnels       map[string]*[]types.TunnelOption
	runtime       map[string]*ssh.Client
	localListener *utils.SyncMapListener
	localConn     *utils.SyncMapConn
	remoteConn    *utils.SyncMapConn
	schema        *schema.Schema
}

func NewClientManager() (manager *ClientManager, err error) {
	manager = new(ClientManager)
	manager.options = make(map[string]*types.SshOption)
	manager.tunnels = make(map[string]*[]types.TunnelOption)
	manager.runtime = make(map[string]*ssh.Client)
	manager.localListener = utils.NewSyncMapListener()
	manager.localConn = utils.NewSyncMapConn()
	manager.remoteConn = utils.NewSyncMapConn()
	manager.schema = schema.New()
	var clientOptions []types.ClientOption
	clientOptions, err = manager.schema.Lists()
	for _, option := range clientOptions {
		var key []byte
		key, err = base64.StdEncoding.DecodeString(option.Key)
		if err != nil {
			return
		}
		var passPhrase []byte
		passPhrase, err = base64.StdEncoding.DecodeString(option.PassPhrase)
		if err != nil {
			return
		}
		err = manager.Put(option.Identity, types.SshOption{
			Host:       option.Host,
			Port:       option.Port,
			Username:   option.Username,
			Password:   option.Password,
			Key:        key,
			PassPhrase: passPhrase,
		})
		if err != nil {
			return
		}
		var tunnels []types.TunnelOption
		for _, tunnelOption := range option.Tunnels {
			tunnels = append(tunnels, types.TunnelOption{
				SrcIp:   tunnelOption.SrcIp,
				SrcPort: tunnelOption.SrcPort,
				DstIp:   tunnelOption.DstIp,
				DstPort: tunnelOption.DstPort,
			})
		}
		if len(tunnels) == 0 {
			continue
		}
		err = manager.Tunnels(option.Identity, tunnels)
		if err != nil {
			return
		}
	}
	return
}

func (c *ClientManager) empty(identity string) error {
	if c.options[identity] == nil || c.runtime[identity] == nil {
		return errors.New("this identity does not exists")
	}
	return nil
}

func (c *ClientManager) GetIdentityCollection() []string {
	var keys []string
	for key := range c.options {
		keys = append(keys, key)
	}
	return keys
}

// Get ssh client information
func (c *ClientManager) GetSshOption(identity string) (option *types.SshOption, err error) {
	if err = c.empty(identity); err != nil {
		return
	}
	option = c.options[identity]
	return
}

func (c *ClientManager) GetRuntime(identity string) (client *ssh.Client, err error) {
	if err = c.empty(identity); err != nil {
		return
	}
	client = c.runtime[identity]
	return
}

func (c *ClientManager) GetTunnelOption(identity string) (option []types.TunnelOption, err error) {
	if err = c.empty(identity); err != nil {
		return
	}
	if c.tunnels[identity] != nil {
		option = *c.tunnels[identity]
	}
	return
}
