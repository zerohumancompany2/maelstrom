package mail

type Subscriber interface {
	Receive() <-chan Mail
}

type ChannelSubscriber struct {
	ch chan Mail
}

func NewChannelSubscriber(bufferSize int) *ChannelSubscriber {
	return &ChannelSubscriber{
		ch: make(chan Mail, bufferSize),
	}
}

func (s *ChannelSubscriber) Receive() <-chan Mail {
	return s.ch
}

func (s *ChannelSubscriber) GetMail() (Mail, bool) {
	select {
	case mail := <-s.ch:
		return mail, true
	default:
		return Mail{}, false
	}
}
