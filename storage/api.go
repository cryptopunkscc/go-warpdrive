package storage

import (
	"github.com/cryptopunkscc/go-warpdrive/proto"
	"io"
	"os"
)

type Offer interface {
	Save(offer proto.Offer)
	Get() proto.Offers
}

type Peer interface {
	Save(peers []proto.Peer)
	Get() proto.Peers
	List() []proto.Peer
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
	Info(uri string) (files []proto.Info, err error)
}
