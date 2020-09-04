package mq

import (
	"elastic-transfer/app/types"
	"github.com/streadway/amqp"
)

type MessageQueue struct {
	types.MqOption
	Amqp *types.AmqpOption
}

func NewMessageQueue(option types.MqOption) (mq *MessageQueue, err error) {
	mq = new(MessageQueue)
	mq.MqOption = option
	if mq.Drive == "amqp" {
		mq.Amqp = new(types.AmqpOption)
		mq.Amqp.Conn, err = amqp.Dial(mq.Url)
		if err != nil {
			return
		}
		mq.Amqp.NotifyConnClose = make(chan *amqp.Error)
		mq.Amqp.Conn.NotifyClose(mq.Amqp.NotifyConnClose)
	}
	return
}

func (c *MessageQueue) Push(topic string, key string, data []byte) (err error) {
	if c.Drive == "amqp" {
		err = c.pushFromAmqp(topic, key, data)
		if err != nil {
			return
		}
	}
	return
}
