package drive

type API interface {
	Publish(topic string, key string, data []byte) (err error)
}
