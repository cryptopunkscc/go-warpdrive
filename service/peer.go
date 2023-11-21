package service

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/go-apphost-jrpc/android/contacts"
	"github.com/cryptopunkscc/go-warpdrive"
	"log"
	"sync"
)

type peer struct {
	mu          *sync.RWMutex
	logger      *log.Logger
	memStorage  warpdrive.PeerStorage
	fileStorage warpdrive.PeerStorage
	client      contacts.Client
}

var _ warpdrive.PeerService = peer{}

func (srv peer) Fetch() {
	connect, err := srv.client.Connect(id.Identity{}, "contacts")
	if err != nil {
		srv.logger.Println("Cannot connect contacts", err)
		return
	}

	c, err := connect.Contacts()
	if err != nil {
		srv.logger.Println("Cannot obtain contacts", err)
		return
	}
	contactList := <-c
	srv.mu.Lock()
	defer srv.mu.Unlock()
	for _, contact := range contactList {
		srv.update(warpdrive.PeerId(contact.Id), "alias", contact.Alias)
	}
	srv.save()
}

func (srv peer) Update(id warpdrive.PeerId, attr string, value string) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.update(id, attr, value)
	srv.save()
}

func (srv peer) update(id warpdrive.PeerId, attr string, value string) {
	mem := srv.memStorage.Get()

	p := mem[id]
	cached := p != nil
	if !cached {
		p = &warpdrive.Peer{Id: id}
		mem[id] = p
	}
	switch attr {
	case "mod":
		p.Mod = value
	case "alias":
		p.Alias = value
	default:
		if cached {
			return
		}
	}
}

func (srv peer) save() {
	var peers []warpdrive.Peer
	mem := srv.memStorage.Get()
	for _, p := range mem {
		peers = append(peers, *p)
	}
	srv.fileStorage.Save(peers)
}

func (srv peer) Get(id warpdrive.PeerId) warpdrive.Peer {
	srv.mu.RLock()
	defer srv.mu.RUnlock()
	p := srv.memStorage.Get()[id]
	if p == nil {
		p = &warpdrive.Peer{
			Id:    id,
			Alias: "",
			Mod:   "",
		}
	}
	return *p
}

func (srv peer) List() (peers []warpdrive.Peer) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()
	srv.Fetch()
	return srv.memStorage.List()
}
