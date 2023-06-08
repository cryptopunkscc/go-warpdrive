package service

import (
	"go-warpdrive/proto"
	"go-warpdrive/storage"
	"sort"
	"sync"
	"time"
)

type offer struct {
	Component
	*proto.Offer
	mu         *sync.RWMutex
	offerSubs  *proto.Subscriptions
	statusSubs *proto.Subscriptions
	file       storage.Offer
	mem        storage.Offer
	incoming   bool
}

var _ proto.OfferService = &offer{}

func (srv *offer) OfferSubscriptions() *proto.Subscriptions {
	return srv.offerSubs
}

func (srv *offer) StatusSubscriptions() *proto.Subscriptions {
	return srv.statusSubs
}

func (srv *offer) Get(id proto.OfferId) (offer *proto.Offer) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()
	offer = srv.mem.Get()[id]
	return
}

func (srv *offer) List() (offers []proto.Offer) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()
	m := srv.mem.Get()
	for _, o := range m {
		offers = append(offers, *o)
	}
	sort.Sort(byCreate(offers))
	return
}

type byCreate []proto.Offer

func (a byCreate) Len() int           { return len(a) }
func (a byCreate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byCreate) Less(i, j int) bool { return a[i].Create < a[j].Create }

func (srv *offer) Add(
	offerId proto.OfferId,
	files []proto.Info,
	peerId proto.PeerId,
) (offer *proto.Offer) {
	offer = &proto.Offer{
		Files:  files,
		Peer:   peerId,
		Create: time.Now().UnixMilli(),
		OfferStatus: proto.OfferStatus{
			Status: proto.StatusAwaiting,
			In:     srv.incoming,
			Id:     offerId,
			Index:  -1,
		},
	}
	srv.dispatch(offer)
	return
}

func (srv *offer) Accept(offer *proto.Offer) {
	offer.Status = proto.StatusAccepted
	srv.dispatch(offer)
}

func (srv *offer) Finish(offer *proto.Offer, err error) {
	if err == nil {
		offer.Index = len(offer.Files)
		offer.Progress = 0
		offer.Status = proto.StatusCompleted
	} else {
		offer.Status = proto.StatusFailed
	}
	srv.dispatch(offer)
}

func (srv *offer) dispatch(offer *proto.Offer) {
	offer.Update = time.Now().UnixMilli()
	srv.Offer = offer
	srv.Offers <- srv
}
