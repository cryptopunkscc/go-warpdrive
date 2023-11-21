package warpdrive

import (
	"context"
	"io"
	"sync"
)

type Service interface {
	Incoming() OfferService
	Outgoing() OfferService
	Peer() PeerService
	File() FileService
	Start(ctx context.Context) <-chan struct{}
	Job() *sync.WaitGroup
}

type PeerService interface {
	Fetch()
	Update(id PeerId, attr string, value string)
	Get(id PeerId) Peer
	List() (peers []Peer)
}

type OfferService interface {
	List() (offers []Offer)
	Get(id OfferId) *Offer
	Add(offerId OfferId, files []Info, peerId PeerId) *Offer
	Accept(offer *Offer)
	Copy(offer *Offer) CopyOffer
	Finish(offer *Offer, err error)
	OfferSubscriptions() *Subscriptions
	StatusSubscriptions() *Subscriptions
}

type CopyOffer interface {
	From(reader io.Reader) (err error)
	To(writer io.Writer) (err error)
}

type FileService interface {
	Info(uri string) (files []Info, err error)
}

const (
	StatusAwaiting  = "awaiting"
	StatusAccepted  = "accepted"
	StatusRejected  = "rejected"
	StatusUpdated   = "updated"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

const (
	PeerModAsk   = ""
	PeerModTrust = "trust"
	PeerModBlock = "block"
)

func FilterOffers(
	srv Service,
	filter Filter,
) (offers []Offer) {
	switch filter {
	case FilterIn:
		offers = append(offers, srv.Incoming().List()...)
	case FilterOut:
		offers = append(offers, srv.Outgoing().List()...)
	case FilterAll:
		offers = append(offers, srv.Incoming().List()...)
		offers = append(offers, srv.Outgoing().List()...)
	}
	return
}
