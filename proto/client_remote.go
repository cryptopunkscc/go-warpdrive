package proto

import (
	"encoding/json"
	"github.com/cryptopunkscc/go-warpdrive"
	"log"
	"net/url"
	"path/filepath"
	"strings"
)

func (c Client) SendOffer(offerId warpdrive.OfferId, files []warpdrive.Info) (code bool, err error) {
	// Request send offer
	err = c.cslq.Encodef("c [c]c", remoteSend, offerId)
	if err != nil {
		err = warpdrive.Error(err, "Cannot send offer with id", offerId)
		return
	}
	shrunken := shrinkPaths(files)
	err = json.NewEncoder(c.conn).Encode(shrunken)
	if err != nil {
		err = warpdrive.Error(err, "Cannot send offer info", offerId)
		return
	}
	// Read result code
	err = c.cslq.Decodef("c", &code)
	if err != nil {
		err = warpdrive.Error(err, "Cannot read result code")
		return
	}
	return
}

// TODO make it bulletproof
func shrinkPaths(in []warpdrive.Info) (out []warpdrive.Info) {
	dir, _ := filepath.Split(in[0].Uri)
	if dir == "" {
		return in
	}
	uri, err := url.Parse(dir)
	if err != nil {
		log.Println("Cannot parse uri", err)
		return in
	}
	if uri.Scheme != "" {
		for _, info := range in {
			uri, err = url.Parse(info.Uri)
			if err != nil {
				log.Println("Cannot parse uri", err)
				return in
			}
			path := strings.Replace(uri.Path, ":", "/", -1)
			_, file := filepath.Split(path)
			info.Uri = file
			out = append(out, info)
		}
	} else {
		for _, info := range in {
			info.Uri = strings.TrimPrefix(info.Uri, dir)
			out = append(out, info)
		}
	}
	return
}

func (c Client) Download(
	offerId warpdrive.OfferId,
	index int,
	offset int64,
) (err error) {
	// Request download
	if err = c.cslq.Encodef("c [c]c q q", remoteDownload, offerId, index, offset); err != nil {
		err = warpdrive.Error(err, "Cannot request download")
	}

	// Read confirmation
	var code byte
	err = c.cslq.Decodef("c", &code)
	if err != nil {
		err = warpdrive.Error(err, "Cannot read confirmation")
		return
	}

	return
}
