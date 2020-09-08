package mq

import "github.com/streadway/amqp"

func (c *MessageQueue) Publish(topic string, key string, data []byte) (err error) {
	if c.Drive == "amqp" {
		err = c.publishFromAmqp(topic, key, data)
		if err != nil {
			return
		}
	}
	return
}

func (c *MessageQueue) publishFromAmqp(exchange string, key string, data []byte) (err error) {
	var channel *amqp.Channel
	channel, err = c.amqp.conn.Channel()
	if err != nil {
		return
	}
	defer channel.Close()
	err = channel.Publish(exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	})
	if err != nil {
		return
	}
	return
}
