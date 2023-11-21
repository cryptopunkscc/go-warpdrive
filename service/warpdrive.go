package service

import (
	"github.com/cryptopunkscc/go-warpdrive"
	"github.com/cryptopunkscc/go-warpdrive/storage/file"
	"github.com/cryptopunkscc/go-warpdrive/storage/memory"
)

var _ warpdrive.Service = &Component{}

func (c *Component) Incoming() warpdrive.OfferService {
	return &offer{
		Component:  *c,
		mu:         &c.Mutex.Incoming,
		mem:        memory.Offer(c.Cache.Incoming),
		file:       file.Incoming(c.Logger, c.RepositoryDir),
		offerSubs:  c.Observers.IncomingOffers,
		statusSubs: c.Observers.IncomingStatus,
		incoming:   true,
	}
}

func (c *Component) Outgoing() warpdrive.OfferService {
	return &offer{
		Component:  *c,
		mu:         &c.Mutex.Outgoing,
		mem:        memory.Offer(c.Cache.Outgoing),
		file:       file.Outgoing(c.Logger, c.RepositoryDir),
		offerSubs:  c.Observers.OutgoingOffers,
		statusSubs: c.Observers.OutgoingStatus,
	}
}

func (c *Component) Peer() warpdrive.PeerService {
	return peer(*c)
}

func (c *Component) File() warpdrive.FileService {
	return c.FileResolver
}
