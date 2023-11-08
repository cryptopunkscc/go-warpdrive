package memory

import (
	"github.com/cryptopunkscc/go-warpdrive/proto"
	"github.com/cryptopunkscc/go-warpdrive/storage"
)

type Peers proto.Peers

var _ storage.Peer = Peers{}

func (r Peers) Save(peers []proto.Peer) {
	for _, peer := range peers {
		p := peer
		r[peer.Id] = &p
	}
}

func (r Peers) Get() proto.Peers {
	return proto.Peers(r)
}

func (r Peers) List() (peers []proto.Peer) {
	p := r.Get()
	for _, peer := range p {
		peers = append(peers, *peer)
	}
	return
}
