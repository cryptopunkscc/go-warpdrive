package service

import (
	"github.com/cryptopunkscc/go-warpdrive/proto"
	"github.com/cryptopunkscc/go-warpdrive/storage/file"
	"github.com/cryptopunkscc/go-warpdrive/storage/memory"
)

type peer Component

var _ proto.PeerService = peer{}

func (srv peer) Fetch() {
	// TODO
	//contactList, err := contacts.Client{Api: srv}.List()
	//if err != nil {
	//	srv.Println("Cannot obtain contacts", err)
	//	return
	//}
	//srv.Mutex.Peers.Lock()
	//defer srv.Mutex.Peers.Unlock()
	//for _, contact := range contactList {
	//	srv.update(warpdrive.PeerId(contact.Id), "alias", contact.Name)
	//}
	//srv.save()
}

func (srv peer) Update(id proto.PeerId, attr string, value string) {
	srv.Mutex.Peers.Lock()
	defer srv.Mutex.Peers.Unlock()
	srv.update(id, attr, value)
	srv.save()
}

func (srv peer) update(id proto.PeerId, attr string, value string) {
	mem := srv.Peers
	p := mem[id]
	cached := p != nil
	if !cached {
		p = &proto.Peer{Id: id}
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
	var peers []proto.Peer
	mem := memory.Peers(srv.Peers).Get()
	for _, p := range mem {
		peers = append(peers, *p)
	}
	file.Peers(srv.Logger, srv.RepositoryDir).Save(peers)
}

func (srv peer) Get(id proto.PeerId) proto.Peer {
	srv.Mutex.Peers.RLock()
	defer srv.Mutex.Peers.RUnlock()
	p := memory.Peers(srv.Peers).Get()[id]
	if p == nil {
		p = &proto.Peer{
			Id:    id,
			Alias: "",
			Mod:   "",
		}
	}
	return *p
}

func (srv peer) List() (peers []proto.Peer) {
	srv.Fetch()
	srv.Mutex.Peers.RLock()
	defer srv.Mutex.Peers.RUnlock()
	return memory.Peers(srv.Peers).List()
}

func (srv peer) Offers() *proto.Subscriptions {
	return srv.IncomingOffers
}
