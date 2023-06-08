package file

import (
	"encoding/gob"
	"go-warpdrive/proto"
	"go-warpdrive/storage"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Incoming(logger *log.Logger, repositoryDir string) storage.Offer {
	o := offers{
		logger,
		filepath.Join(repositoryDir, "incoming"),
	}
	o.Init()
	return o
}

func Outgoing(logger *log.Logger, repositoryDir string) storage.Offer {
	o := offers{
		logger,
		filepath.Join(repositoryDir, "outgoing"),
	}
	o.Init()
	return o
}

type offers struct {
	*log.Logger
	dir string
}

var _ storage.Offer = offers{}

func (r offers) Init() {
	_ = os.MkdirAll(r.normalizePath(""), 0700)
}

func (r offers) Save(offer proto.Offer) {
	path := r.normalizePath(string(offer.Id))

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		r.Panicln("cannot create file for incoming offer", err)
	}
	err = gob.NewEncoder(file).Encode(offer)
	if err != nil {
		r.Panicln("cannot write offer to file", err)
	}
	err = file.Close()
	if err != nil {
		r.Println("cannot close offer file", path, err)
	}
}

func (r offers) Get() proto.Offers {
	offers := make(proto.Offers)
	dir := r.normalizePath("")
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		normalizedPath := r.normalizePath(path)
		file, err := os.Open(normalizedPath)
		if err != nil {
			r.Println("cannot open", err)
			return nil
		}
		id := proto.OfferId(info.Name())
		offer := &proto.Offer{}
		err = gob.NewDecoder(file).Decode(offer)
		if err != nil {
			r.Println("cannot decode", path, err)
			return nil
		}
		offers[id] = offer
		return nil
	})
	if err != nil {
		r.Println("Cannot list incoming offers", err)
	}
	return offers
}

func (r offers) normalizePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return filepath.Join(r.dir, path)
}