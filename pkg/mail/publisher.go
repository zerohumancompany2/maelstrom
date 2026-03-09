package mail

import "time"

type Publisher interface {
	Publish(mail Mail) (Ack, error)
}

type RouterPublisher struct {
	router *MailRouter
}

func NewRouterPublisher(router *MailRouter) *RouterPublisher {
	return &RouterPublisher{router: router}
}

func (p *RouterPublisher) Publish(mail Mail) (Ack, error) {
	err := p.router.Route(mail)
	return Ack{
		CorrelationID: mail.CorrelationID,
		DeliveredAt:   time.Now(),
	}, err
}
