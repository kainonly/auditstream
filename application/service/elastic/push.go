package elastic

import "elastic-transfer/application/common/actions"

func (c *Elastic) Push(identity string, content []byte) (err error) {
	if c.Pipes.Empty(identity) {
		return NotExists
	}
	pipe := c.Pipes.Get(identity)
	if err = actions.Push(c.Client, pipe.Index, content); err != nil {
		if err = c.Queue.Publish(pipe.Topic, pipe.Key, content); err != nil {
			return
		}
	}
	return
}
