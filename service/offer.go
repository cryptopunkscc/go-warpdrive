package service

import (
	"github.com/cryptopunkscc/go-warpdrive"
	"log"
	"sort"
	"sync"
	"time"
)

var _ warpdrive.OfferService = &offerService{}

type offerService struct {
	*warpdrive.Offer
	*log.Logger

	mu              *sync.RWMutex
	offersBroadcast *warpdrive.Broadcast[warpdrive.Offer]
	statusBroadcast *warpdrive.Broadcast[warpdrive.OfferStatus]

	offerFileStorage warpdrive.OfferStorage
	offerMemStorage  warpdrive.OfferStorage

	resolver    warpdrive.FileResolver
	fileStorage warpdrive.FileStorage
	peerStorage warpdrive.PeerStorage
	changes     chan *offerService
	incoming    bool
}

func newOfferService(
	current *warpdrive.Offer,
	logger *log.Logger,
	mu *sync.RWMutex,
	offerFileStorage warpdrive.OfferStorage,
	offerMemStorage warpdrive.OfferStorage,
	resolver warpdrive.FileResolver,
	fileStorage warpdrive.FileStorage,
	peerStorage warpdrive.PeerStorage,
	changes chan *offerService,
	incoming bool,
) *offerService {
	return &offerService{
		Offer:            current,
		Logger:           logger,
		mu:               mu,
		offersBroadcast:  &warpdrive.Broadcast[warpdrive.Offer]{},
		statusBroadcast:  &warpdrive.Broadcast[warpdrive.OfferStatus]{},
		offerFileStorage: offerFileStorage,
		offerMemStorage:  offerMemStorage,
		resolver:         resolver,
		fileStorage:      fileStorage,
		peerStorage:      peerStorage,
		changes:          changes,
		incoming:         incoming,
	}
}

func (srv *offerService) OfferBroadcast() *warpdrive.Broadcast[warpdrive.Offer] {
	return srv.offersBroadcast
}

func (srv *offerService) StatusBroadcast() *warpdrive.Broadcast[warpdrive.OfferStatus] {
	return srv.statusBroadcast
}

func (srv *offerService) Get(id warpdrive.OfferId) (offer *warpdrive.Offer) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()
	offer = srv.offerMemStorage.GetMap()[id]
	return
}

func (srv *offerService) List() (offers []warpdrive.Offer) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()
	m := srv.offerMemStorage.GetMap()
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

func (srv *offerService) Add(
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
	srv.update(offer)
	return
}

func (srv *offerService) Accept(offer *warpdrive.Offer) {
	offer.Status = warpdrive.StatusAccepted
	srv.update(offer)
}

func (srv *offerService) Finish(offer *warpdrive.Offer, err error) {
	if err == nil {
		offer.Index = len(offer.Files)
		offer.Progress = 0
		offer.Status = warpdrive.StatusCompleted
	} else {
		offer.Status = warpdrive.StatusFailed
	}
	srv.update(offer)
}

func (srv *offerService) update(offer *warpdrive.Offer) {
	offer.Update = time.Now().UnixMilli()
	srv.Offer = offer
	srv.changes <- srv
}
