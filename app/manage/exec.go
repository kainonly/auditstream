package manage

import "sync"

func (c *ClientManager) Exec(identity string, cmd string) (output []byte, err error) {
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
