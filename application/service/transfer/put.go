package transfer

import "elastic-transfer/config/options"

func (c *Transfer) Put(option options.PipeOption) (err error) {
	c.Pipes.Put(option.Identity, &option)
	return c.Schema.Update(option)
}
