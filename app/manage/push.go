package manage

import (
	"elastic-transfer/app/actions"
	"time"
)

func (c *ElasticManager) Push(identity string, data []byte) (err error) {
	if err = c.empty(identity); err != nil {
		return
	}
	pipe := c.pipes[identity]
	time.Sleep(time.Microsecond * time.Duration(c.runtime))
	err = c.pool.Submit(func() {
		c.runtime++
		err = actions.Push(c.client, pipe.Index, data)
		if err != nil {
			c.mq.Publish(pipe.Topic, pipe.Key, data)
		}
		c.runtime--
	})
	return
}
