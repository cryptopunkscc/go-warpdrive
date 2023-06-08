package warpdrived

import (
	"fmt"
	"go-warpdrive/proto"
	"go-warpdrive/service"
	"go-warpdrive/storage/file"
	"log"
	"os"
	"path/filepath"
	"sync"
)

func setupCore(c *service.Component) {
	// Defaults
	c.Logger = log.Default()
	c.Job = &sync.WaitGroup{}
	if c.Sys == nil {
		c.Sys = &service.Sys{}
	}
	c.Channel = &service.Channel{}
	c.Cache = &service.Cache{
		Mutex:    &service.Mutex{},
		Incoming: proto.Offers{},
		Outgoing: proto.Offers{},
		Peers:    proto.Peers{},
	}
	c.Observers = &service.Observers{
		IncomingOffers: proto.NewSubscriptions(),
		IncomingStatus: proto.NewSubscriptions(),
		OutgoingOffers: proto.NewSubscriptions(),
		OutgoingStatus: proto.NewSubscriptions(),
	}

	// Platform
	if c.Platform == "" {
		c.Platform = service.PlatformDefault
	}

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
	if c.Sys.Notify == nil {
		log.Println("WARNING: using stub notify") // TODO make better log
		c.Sys.Notify = func(_ []service.Notification) {}
	}

	// Resolver
	if c.FileResolver == nil {
		log.Println("WARNING: using default file.Resolver") // TODO make better log
		c.FileResolver = &file.Resolver{}
	}
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
