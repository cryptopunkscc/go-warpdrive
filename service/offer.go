package service

import (
	"github.com/cryptopunkscc/go-warpdrive"
	"sort"
	"sync"
	"time"
)

type offer struct {
	Component
	*warpdrive.Offer
	mu         *sync.RWMutex
	offerSubs  *warpdrive.Subscriptions
	statusSubs *warpdrive.Subscriptions
	file       warpdrive.OfferStorage
	mem        warpdrive.OfferStorage
	incoming   bool
}

var _ warpdrive.OfferService = &offer{}

func (srv *offer) OfferSubscriptions() *warpdrive.Subscriptions {
	return srv.offerSubs
}

func (srv *offer) StatusSubscriptions() *warpdrive.Subscriptions {
	return srv.statusSubs
}

func (srv *offer) Get(id warpdrive.OfferId) (offer *warpdrive.Offer) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()
	offer = srv.mem.Get()[id]
	return
}

func (srv *offer) List() (offers []warpdrive.Offer) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()
	m := srv.mem.Get()
	for _, o := range m {
		offers = append(offers, *o)
	}
	sort.Sort(byCreate(offers))
	return
}

type byCreate []warpdrive.Offer

func (a byCreate) Len() int           { return len(a) }
func (a byCreate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byCreate) Less(i, j int) bool { return a[i].Create < a[j].Create }

func (srv *offer) Add(
	offerId warpdrive.OfferId,
	files []warpdrive.Info,
	peerId warpdrive.PeerId,
) (offer *warpdrive.Offer) {
	offer = &warpdrive.Offer{
		Files:  files,
		Peer:   peerId,
		Create: time.Now().UnixMilli(),
		OfferStatus: warpdrive.OfferStatus{
			Status: warpdrive.StatusAwaiting,
			In:     srv.incoming,
			Id:     offerId,
			Index:  -1,
		},
	}
	srv.dispatch(offer)
	return
}

func (srv *offer) Accept(offer *warpdrive.Offer) {
	offer.Status = warpdrive.StatusAccepted
	srv.dispatch(offer)
}

func (srv *offer) Finish(offer *warpdrive.Offer, err error) {
	if err == nil {
		offer.Index = len(offer.Files)
		offer.Progress = 0
		offer.Status = warpdrive.StatusCompleted
	} else {
		offer.Status = warpdrive.StatusFailed
	}
	srv.dispatch(offer)
}

func (srv *offer) dispatch(offer *warpdrive.Offer) {
	offer.Update = time.Now().UnixMilli()
	srv.Offer = offer
	srv.Offers <- srv
}
