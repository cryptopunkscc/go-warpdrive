package service

import (
	"github.com/cryptopunkscc/go-warpdrive/proto"
	"github.com/cryptopunkscc/go-warpdrive/storage/file"
	"github.com/cryptopunkscc/go-warpdrive/storage/memory"
)

var _ proto.Service = Component{}

func (w Component) Incoming() proto.OfferService {
	c := w
	return &offer{
		Component:  c,
		mu:         &c.Mutex.Incoming,
		mem:        memory.Offer(c.Cache.Incoming),
		file:       file.Incoming(c.Logger, c.RepositoryDir),
		offerSubs:  c.Observers.IncomingOffers,
		statusSubs: c.Observers.IncomingStatus,
		incoming:   true,
	}
}

func (w Component) Outgoing() proto.OfferService {
	c := w
	return &offer{
		Component:  c,
		mu:         &c.Mutex.Outgoing,
		mem:        memory.Offer(c.Cache.Outgoing),
		file:       file.Outgoing(c.Logger, c.RepositoryDir),
		offerSubs:  c.Observers.OutgoingOffers,
		statusSubs: c.Observers.OutgoingStatus,
	}
}

func (w Component) Peer() proto.PeerService {
	return peer(w)
}

func (w Component) File() proto.FileService {
	return w.FileResolver
}
