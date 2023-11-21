package warpdrive

type CreateNotify func() Notify

type Notify func([]Notification)

type Notification struct {
	Peer
	Offer
	*Info
}
