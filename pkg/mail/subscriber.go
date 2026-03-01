package mail

type Subscriber interface {
	Subscribe(address string) (<-chan Mail, error)
	Unsubscribe(address string, ch <-chan Mail) error
}
