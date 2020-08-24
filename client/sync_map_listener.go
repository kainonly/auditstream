package client

import (
	"net"
	"sync"
)

type syncMapListener struct {
	sync.RWMutex
	Map map[string]map[string]*net.Listener
}

func newSyncMapListener() *syncMapListener {
	c := new(syncMapListener)
	c.Map = make(map[string]map[string]*net.Listener)
	return c
}

func (c *syncMapListener) Clear(identity string) {
	delete(c.Map, identity)
}

func (c *syncMapListener) Set(identity string, addr string, listener *net.Listener) {
	c.Lock()
	if c.Map[identity] == nil {
		c.Map[identity] = make(map[string]*net.Listener)
	}
	c.Map[identity][addr] = listener
	c.Unlock()
}
