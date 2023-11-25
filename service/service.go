package service

import (
	"context"
	"github.com/cryptopunkscc/go-warpdrive"
	"log"
	"sync"
)

type service struct {
	incoming offerService
	outgoing offerService
	updates  offerUpdates
	peer     peer
	file     warpdrive.FileResolver
	job      *sync.WaitGroup
	done     <-chan struct{}
}

func Start(
	ctx context.Context,
	logger *log.Logger,
	notify Notify,
	storageFactory warpdrive.StorageFactory,
) warpdrive.Service {
	job := &sync.WaitGroup{}
	inMu := &sync.RWMutex{}
	outMu := &sync.RWMutex{}
	changes := make(chan *offerService, 1024)
	peers := storageFactory.Peer()
	inOffers := storageFactory.Offer("incoming")
	outOffers := storageFactory.Offer("outgoing")
	srv := service{
		job:  job,
		file: storageFactory.Resolver(),
		updates: offerUpdates{
			channel: changes,
			log:     logger,
			inMu:    inMu,
			outMu:   outMu,
			job:     job,
			notify:  notify,
		},
		incoming: offerService{
			Logger:           logger,
			mu:               inMu,
			offersBroadcast:  warpdrive.NewBroadcast[warpdrive.Offer](),
			statusBroadcast:  warpdrive.NewBroadcast[warpdrive.OfferStatus](),
			offerFileStorage: inOffers,
			offerMemStorage:  storageFactory.OfferCache(inOffers.GetMap()),
			resolver:         storageFactory.Resolver(),
			fileStorage:      storageFactory.File(),
			peerStorage:      peers,
			changes:          changes,
			incoming:         true,
		},
		outgoing: offerService{
			Logger:           logger,
			mu:               outMu,
			offersBroadcast:  warpdrive.NewBroadcast[warpdrive.Offer](),
			statusBroadcast:  warpdrive.NewBroadcast[warpdrive.OfferStatus](),
			offerFileStorage: outOffers,
			offerMemStorage:  storageFactory.OfferCache(outOffers.GetMap()),
			resolver:         storageFactory.Resolver(),
			peerStorage:      peers,
			changes:          changes,
			incoming:         false,
		},
		peer: peer{
			mu:          &sync.RWMutex{},
			memStorage:  storageFactory.PeerCache(),
			fileStorage: storageFactory.Peer(),
		},
	}
	srv.done = srv.updates.Start(ctx)
	return srv
}

func (s service) Incoming() warpdrive.OfferService {
	return &s.incoming
}

func (s service) Outgoing() warpdrive.OfferService {
	return &s.outgoing
}

func (s service) Peer() warpdrive.PeerService {
	return s.peer
}

func (s service) File() warpdrive.FileService {
	return s.file
}

func (s service) Start(ctx context.Context) <-chan struct{} {
	return s.updates.Start(ctx)
}

func (s service) Done() <-chan struct{} {
	return s.done
}

func (s service) Job() *sync.WaitGroup {
	return s.job
}
