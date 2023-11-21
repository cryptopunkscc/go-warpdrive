package storage

import (
	"github.com/cryptopunkscc/go-warpdrive"
	"github.com/cryptopunkscc/go-warpdrive/storage/file"
	"github.com/cryptopunkscc/go-warpdrive/storage/memory"
	"log"
	"path"
)

type Factory struct {
	logger       *log.Logger
	cacheDir     string
	filesDir     string
	FileResolver warpdrive.FileResolver
}

func NewFactory(
	logger *log.Logger,
	cacheDir string,
	filesDir string,
) *Factory {
	return &Factory{
		logger:       logger,
		cacheDir:     cacheDir,
		filesDir:     filesDir,
		FileResolver: file.NewResolver(),
	}
}

func (f Factory) OfferCache(offers warpdrive.Offers) warpdrive.OfferStorage {
	return memory.Offer(offers)
}

func (f Factory) Offer(dir string) warpdrive.OfferStorage {
	return file.NewOffersStorage(f.logger, path.Join(f.cacheDir, dir))
}

func (f Factory) PeerCache() warpdrive.PeerStorage {
	return memory.Peers{}
}

func (f Factory) Peer() warpdrive.PeerStorage {
	return file.NewPeersStorage(f.logger, path.Join("peers"))
}

func (f Factory) File() warpdrive.FileStorage {
	return file.NewStorage(f.filesDir)
}

func (f Factory) Resolver() warpdrive.FileResolver {
	return f.FileResolver
}
