package file

import (
	"encoding/gob"
	"go-warpdrive/proto"
	"go-warpdrive/storage"
	"log"
	"os"
	"path/filepath"
)

func Peers(logger *log.Logger, repositoryDir string) storage.Peer {
	return peers{
		logger,
		filepath.Join(repositoryDir, "peers"),
	}
}

type peers struct {
	*log.Logger
	path string
}

func (r peers) Save(peers []proto.Peer) {
	file, err := os.OpenFile(r.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		r.Panicln("cannot open peers file", err)
	}
	err = gob.NewEncoder(file).Encode(peers)
	if err != nil {
		r.Panicln("cannot write peers", err)
	}
}

func (r peers) Get() (peers proto.Peers) {
	list := r.List()
	peers = proto.Peers{}
	for _, peer := range list {
		p := peer
		peers[peer.Id] = &p
	}
	return
}

func (r peers) List() (peers []proto.Peer) {
	file, err := os.Open(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		r.Panicln("cannot open peers file", err)
	}
	err = gob.NewDecoder(file).Decode(&peers)
	if err != nil {
		r.Panicln("cannot read peers file", err)
	}
	return
}
