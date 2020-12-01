package drive

import (
	"github.com/streadway/amqp"
	"log"
	"time"
)

type AMQPDrive struct {
	option          AMQPOption
	conn            *amqp.Connection
	notifyConnClose chan *amqp.Error
	API
}

type AMQPOption struct {
	Url string `yaml:"url"`
}

func InitializeAMQP(option AMQPOption) (c *AMQPDrive, err error) {
	c = new(AMQPDrive)
	c.option = option
	if c.conn, err = amqp.Dial(option.Url); err != nil {
		return
	}
	c.notifyConnClose = make(chan *amqp.Error)
	c.conn.NotifyClose(c.notifyConnClose)
	go c.listenConn()
	return
}

func (c *AMQPDrive) listenConn() {
	select {
	case <-c.notifyConnClose:
		log.Println("AMQP connection has been disconnected")
		c.reconnected()
	}
}

func (c *AMQPDrive) reconnected() {
	var err error
	count := 0
	for {
		time.Sleep(time.Second * 5)
		count++
		log.Println("Trying to reconnect:", count)
		if c.conn, err = amqp.Dial(c.option.Url); err != nil {
			log.Println(err)
			continue
		}
		c.notifyConnClose = make(chan *amqp.Error)
		c.conn.NotifyClose(c.notifyConnClose)
		go c.listenConn()
		log.Println("Attempt to reconnect successfully")
		break
	}
}

func (c *AMQPDrive) Publish(topic string, key string, data []byte) (err error) {
	var channel *amqp.Channel
	if channel, err = c.conn.Channel(); err != nil {
		return
	}
	defer channel.Close()
	if err = channel.Publish(topic, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	}); err != nil {
		return
	}
	return
}
