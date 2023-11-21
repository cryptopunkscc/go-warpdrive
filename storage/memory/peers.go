package memory

import (
	"github.com/cryptopunkscc/go-warpdrive"
)

type Peers warpdrive.Peers

var _ warpdrive.PeerStorage = Peers{}

func (r Peers) Save(peers []warpdrive.Peer) {
	for _, peer := range peers {
		p := peer
		r[peer.Id] = &p
	}
}

func (r Peers) Get() warpdrive.Peers {
	return warpdrive.Peers(r)
}

func (r Peers) List() (peers []warpdrive.Peer) {
	p := r.Get()
	for _, peer := range p {
		peers = append(peers, *peer)
	}
	return
}
