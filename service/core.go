package service

import (
	"github.com/cryptopunkscc/go-warpdrive/adapter"
	"github.com/cryptopunkscc/go-warpdrive/proto"
	"github.com/cryptopunkscc/go-warpdrive/storage"
	"log"
	"sync"
)

type Component struct {
	Config
	*log.Logger
	*Sys
	*Cache
	*Observers
	*Channel
	adapter.Api
	storage.FileResolver
	Job *sync.WaitGroup
}

type Config struct {
	RepositoryDir string
	StorageDir    string
	Platform      string
}

type Sys struct {
	Notify Notify
}

type Cache struct {
	*Mutex
	Incoming proto.Offers
	Outgoing proto.Offers
	Peers    proto.Peers
}

type Mutex struct {
	Incoming sync.RWMutex
	Outgoing sync.RWMutex
	Peers    sync.RWMutex
}

type Observers struct {
	IncomingOffers *proto.Subscriptions
	IncomingStatus *proto.Subscriptions
	OutgoingOffers *proto.Subscriptions
	OutgoingStatus *proto.Subscriptions
}

type Channel struct {
	Offers chan<- interface{}
}

type Notify func([]Notification)

type Notification struct {
	proto.Peer
	proto.Offer
	*proto.Info
}

const (
	PlatformDesktop = "desktop"
	PlatformAndroid = "android"
	PlatformDefault = PlatformDesktop
)
