package manage

import (
	"encoding/base64"
	"microtools-gossh/app/actions"
	"microtools-gossh/app/types"
	"sync"
)

func (c *ClientManager) Put(identity string, option types.SshOption) (err error) {
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
		c.runtime[identity], err = actions.Connect(option)
		if err != nil {
			return
		}
		go c.runtime[identity].Wait()
	}()
	wg.Wait()
	return c.schema.Update(types.ClientOption{
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
