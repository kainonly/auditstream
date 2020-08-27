package utils

import (
	"net"
	"sync"
)

type SyncMapListener struct {
	sync.RWMutex
	Map map[string]map[string]*net.Listener
}

func NewSyncMapListener() *SyncMapListener {
	c := new(SyncMapListener)
	c.Map = make(map[string]map[string]*net.Listener)
	return c
}

func (c *SyncMapListener) Clear(identity string) {
	delete(c.Map, identity)
}

func (c *SyncMapListener) Get(identity string, addr string) *net.Listener {
	c.RLock()
	listener := c.Map[identity][addr]
	c.RUnlock()
	return listener
}

func (c *SyncMapListener) Set(identity string, addr string, listener *net.Listener) {
	c.Lock()
	if c.Map[identity] == nil {
		c.Map[identity] = make(map[string]*net.Listener)
	}
	c.Map[identity][addr] = listener
	c.Unlock()
}
