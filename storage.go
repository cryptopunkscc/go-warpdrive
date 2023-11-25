package warpdrive

import (
	"io"
	"os"
)

type StorageFactory interface {
	OfferCache(offers Offers) OfferStorage
	Offer(dir string) OfferStorage
	PeerCache() PeerStorage
	Peer() PeerStorage
	File() FileStorage
	Resolver() FileResolver
}

type OfferStorage interface {
	Save(offer Offer)
	GetMap() Offers
}
type Offers map[OfferId]*Offer

type PeerStorage interface {
	Save(peers []Peer)
	Get() Peers
	List() []Peer
}

type FileStorage interface {
	IsExist(err error) bool
	MkDir(path string, perm os.FileMode) error
	FileWriter(path string, perm os.FileMode, offset int64) (io.WriteCloser, error)
}

// FileResolver provides file reader for uri.
// Required for platforms where direct access to the file system is restricted.
type FileResolver interface {
	Reader(uri string, offset int64) (io.ReadCloser, error)
	Info(uri string) (files []Info, err error)
}
