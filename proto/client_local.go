package proto

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/go-warpdrive"
	"io"
)

func (c Client) CreateOffer(peerId warpdrive.PeerId, filePath string) (os warpdrive.OfferStatus, err error) {
	// Request create offer
	err = c.cslq.Encodef("c [c]c [c]c", localCreateOffer, peerId, filePath)
	if err != nil {
		err = warpdrive.Error(err, "Cannot create offer")
		return
	}
	// Read result
	var accepted bool
	err = c.cslq.Decodef("[c]c c", &os.Id, &accepted)
	if err != nil {
		err = warpdrive.Error(err, "Cannot read create offer results")
		return
	}
	if accepted {
		os.Status = warpdrive.StatusAccepted
	} else {
		os.Status = warpdrive.StatusAwaiting
	}
	return
}

func (c Client) AcceptOffer(id warpdrive.OfferId) (err error) {
	// Request accept offer
	err = c.cslq.Encodef("c [c]c", localAcceptOffer, id)
	if err != nil {
		err = warpdrive.Error(err, "Cannot request accept")
		return
	}
	// Read OK
	var code byte
	err = c.cslq.Decodef("c", &code)
	if err != nil {
		err = warpdrive.Error(err, "Cannot read ok")
		return
	}
	return
}

func (c Client) ListOffers(filter warpdrive.Filter) (offers []warpdrive.Offer, err error) {
	// Request list offers
	err = c.cslq.Encodef("c c", localListOffers, filter)
	if err != nil {
		err = warpdrive.Error(err, "Cannot request offer list")
		return
	}
	// Receive offers
	if err = json.NewDecoder(c.conn).Decode(&offers); err != nil {
		err = warpdrive.Error(err, "Cannot read offers")
		return
	}
	return
}

func (c Client) ListPeers() (peers []warpdrive.Peer, err error) {
	// Request peers
	err = c.cslq.Encodef("c", localListPeers)
	if err != nil {
		err = warpdrive.Error(err, "Cannot request peers")
		return
	}
	// Read peers
	if err = json.NewDecoder(c.conn).Decode(&peers); err != nil {
		err = warpdrive.Error(err, "Cannot read peers")
		return
	}
	return
}

func (c Client) ListenStatus(filter warpdrive.Filter) (status <-chan warpdrive.OfferStatus, err error) {
	// Request status
	err = c.cslq.Encodef("c c", localListenStatus, filter)
	if err != nil {
		err = warpdrive.Error(err, "Cannot request status")
		return
	}
	statChan := make(chan warpdrive.OfferStatus)
	status = statChan
	go func(conn io.ReadWriteCloser, status chan warpdrive.OfferStatus) {
		defer close(status)
		dec := json.NewDecoder(conn)
		files := &warpdrive.OfferStatus{}
		c.log.Println("Start listening status")
		for {
			err := dec.Decode(files)
			if err != nil {
				if fmt.Sprint(errors.Unwrap(err)) == "use of closed network connection" {
					err = nil
				}
				c.log.Println(warpdrive.Error(err, "Finish listening offer status"))
				return
			}
			status <- *files
		}
	}(c.conn, statChan)
	return
}

func (c Client) ListenOffers(filter warpdrive.Filter) (out <-chan warpdrive.Offer, err error) {
	// Request subscribe
	err = c.cslq.Encodef("c c", localListenOffers, filter)
	if err != nil {
		err = warpdrive.Error(err, "Cannot request listen offers")
		return
	}
	offers := make(chan warpdrive.Offer)
	out = offers
	go func() {
		defer close(offers)
		c.log.Println("Start listening offers")
		dec := json.NewDecoder(c.conn)
		for {
			offer := &warpdrive.Offer{}
			err := dec.Decode(offer)
			if err != nil {
				if fmt.Sprint(errors.Unwrap(err)) == "use of closed network connection" {
					err = nil
				}
				c.log.Print(warpdrive.Error(err, "Finish listening new offers"))
				return
			}
			offers <- *offer
		}
	}()
	return
}

func (c Client) UpdatePeer(
	peerId warpdrive.PeerId,
	attr string,
	value string,
) (err error) {
	// Request peer update
	err = c.cslq.Encodef("c", localUpdatePeer)
	if err != nil {
		err = warpdrive.Error(err, "Cannot update peer")
		return
	}
	// Send peers to update
	req := []string{string(peerId), attr, value}
	err = json.NewEncoder(c.conn).Encode(req)
	if err != nil {
		err = warpdrive.Error(err, "Cannot send peer update")
		return
	}
	// Wait for OK
	var code byte
	err = c.cslq.Decodef("c", &code)
	if err != nil {
		err = warpdrive.Error(err, "Cannot read ok")
		return
	}
	return
}
