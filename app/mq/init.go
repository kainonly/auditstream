package mq

import (
	"elastic-transfer/app/types"
)

type MessageQueue struct {
	types.MqOption
	amqp *AmqpDrive
}

func NewMessageQueue(option types.MqOption) (mq *MessageQueue, err error) {
	mq = new(MessageQueue)
	mq.MqOption = option
	if mq.Drive == "amqp" {
		mq.amqp, err = NewAmqpDrive(mq.Url)
		if err != nil {
			return
		}
	}
	return
}
