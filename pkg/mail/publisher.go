package mail

type Publisher interface {
	Publish(mail Mail) (Ack, error)
}
