package service

import (
	"fmt"
	"github.com/cryptopunkscc/go-warpdrive"
	"github.com/cryptopunkscc/go-warpdrive/storage/file"
	"log"
	"os"
	"path/filepath"
	"sync"
)

func NewComponent() (c Component) {
	// Defaults
	c.Logger = log.Default()
	c.job = &sync.WaitGroup{}
	if c.Sys == nil {
		c.Sys = &Sys{}
	}
	c.Channel = &Channel{}
	c.Cache = &Cache{
		Mutex:    &Mutex{},
		Incoming: warpdrive.Offers{},
		Outgoing: warpdrive.Offers{},
		Peers:    warpdrive.Peers{},
	}
	c.Observers = &Observers{
		IncomingOffers: warpdrive.NewSubscriptions(),
		IncomingStatus: warpdrive.NewSubscriptions(),
		OutgoingOffers: warpdrive.NewSubscriptions(),
		OutgoingStatus: warpdrive.NewSubscriptions(),
	}

	// Platform
	c.Platform = PlatformDefault

	// Storage
	c.StorageDir = storageDir()

	// Repository
	if c.RepositoryDir == "" {
		c.RepositoryDir = repositoryDir()
	}

	// Peers
	c.Cache.Peers = file.Peers(c.Logger, c.RepositoryDir).Get()

	// Offers
	c.Cache.Incoming = file.Incoming(c.Logger, c.RepositoryDir).Get()
	c.Cache.Outgoing = file.Outgoing(c.Logger, c.RepositoryDir).Get()

	// Notify
	c.Sys.Notify = func(_ []Notification) {}

	// Resolver
	c.FileResolver = file.Resolver{}
	return
}

func storageDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	dir := filepath.Join(home, "warpdrive", "received")
	return dir
}

func repositoryDir() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("error fetching user's config dir:", err)
		os.Exit(0)
	}
	dir := filepath.Join(cfgDir, "warpdrive")
	os.MkdirAll(dir, 0700)
	return dir
}
