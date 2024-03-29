package file

import (
	"encoding/gob"
	"github.com/cryptopunkscc/go-warpdrive"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type offers struct {
	*log.Logger
	dir string
}

func NewOffersStorage(logger *log.Logger, dir string) warpdrive.OfferStorage {
	o := offers{
		logger,
		dir,
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		panic(err)
	}
	return o
}

func (r offers) Save(offer warpdrive.Offer) {
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

func (r offers) GetMap() warpdrive.Offers {
	offers := make(warpdrive.Offers)
	dir := r.normalizePath("")
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		normalizedPath := r.normalizePath(path)
		file, err := os.Open(normalizedPath)
		if err != nil {
			r.Println("cannot open:", err)
			return nil
		}
		id := warpdrive.OfferId(info.Name())
		offer := &warpdrive.Offer{}
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
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, r.dir) {
		return path
	}
	return filepath.Join(r.dir, path)
}
