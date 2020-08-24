package client

import (
	"net"
	"sync"
)

type syncMapConn struct {
	sync.RWMutex
	Map map[string]map[string]*net.Conn
}

func newSyncMapConn() *syncMapConn {
	listener := new(syncMapConn)
	listener.Map = make(map[string]map[string]*net.Conn)
	return listener
}

func (c *syncMapConn) Clear(identity string) {
	delete(c.Map, identity)
}

func (c *syncMapConn) Get(identity string, addr string) *net.Conn {
	c.RLock()
	conn := c.Map[identity][addr]
	c.RUnlock()
	return conn
}

func (c *syncMapConn) Set(identity string, addr string, conn *net.Conn) {
	c.Lock()
	if c.Map[identity] == nil {
		c.Map[identity] = make(map[string]*net.Conn)
	}
	c.Map[identity][addr] = conn
	c.Unlock()
}
