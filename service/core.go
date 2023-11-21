package service

import (
	"github.com/cryptopunkscc/go-warpdrive"
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
	warpdrive.FileResolver
	job *sync.WaitGroup
}

func (c *Component) Job() *sync.WaitGroup {
	return c.job
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
	Incoming warpdrive.Offers
	Outgoing warpdrive.Offers
	Peers    warpdrive.Peers
}

type Mutex struct {
	Incoming sync.RWMutex
	Outgoing sync.RWMutex
	Peers    sync.RWMutex
}

type Observers struct {
	IncomingOffers *warpdrive.Subscriptions
	IncomingStatus *warpdrive.Subscriptions
	OutgoingOffers *warpdrive.Subscriptions
	OutgoingStatus *warpdrive.Subscriptions
}

type Channel struct {
	Offers chan<- interface{}
}

type Notify func([]Notification)

type Notification struct {
	warpdrive.Peer
	warpdrive.Offer
	*warpdrive.Info
}

const (
	PlatformDesktop = "desktop"
	PlatformAndroid = "android"
	PlatformDefault = PlatformDesktop
)
