package manage

import "elastic-transfer/app/actions"

func (c *ElasticManager) Push(identity string, data []byte) (err error) {
	if err = c.empty(identity); err != nil {
		return
	}
	pipe := c.pipes[identity]
	go func() {
		err = actions.Push(c.client, pipe.Index, data)
		if err != nil {
			c.mq.Publish(pipe.Topic, pipe.Key, data)
		}
	}()
	return
}
