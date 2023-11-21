package warpdrive

import (
	"io"
	"os"
)

type OfferStorage interface {
	Save(offer Offer)
	Get() Offers
}

type PeerStorage interface {
	Save(peers []Peer)
	Get() Peers
	List() []Peer
}

type File interface {
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
