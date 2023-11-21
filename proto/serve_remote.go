package proto

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/go-warpdrive"
	"time"
)

func (d Dispatcher) Ping(any) (err error) {
	finish := make(chan struct{})
	defer close(finish)
	go func() {
		select {
		case <-d.ctx.Done():
		case <-finish:
		}
		_ = d.conn.Close()
	}()
	for {
		var code byte
		if err = cslq.Decode(d.conn, "c", &code); err != nil {
			err = warpdrive.Error(err, "Cannot read ping")
			return
		}
		if err = d.cslq.Encodef("c", code); err != nil {
			err = warpdrive.Error(err, "Cannot write ping")
			return
		}
		if code == 0 {
			return
		}
	}
}

func (d Dispatcher) Receive(any) (err error) {
	peerId := warpdrive.PeerId(d.callerId)
	peer := d.srv.Peer().Get(peerId)
	// Check if peer is blocked
	if peer.Mod == warpdrive.PeerModBlock {
		d.conn.Close()
		d.log.Println("Blocked request from", peerId)
		return
	}
	// Read file offer id
	var offerId warpdrive.OfferId
	err = d.cslq.Decodef("[c]c", &offerId)
	if err != nil {
		err = warpdrive.Error(err, "Cannot read offer id")
		return
	}
	// Read files request
	var files []warpdrive.Info
	err = json.NewDecoder(d.conn).Decode(&files)
	if err != nil {
		err = warpdrive.Error(err, "Cannot read files for offer", offerId)
		return
	}
	// Store incoming offer
	d.srv.Incoming().Add(offerId, files, peerId)
	// Auto accept offer if peer is trusted
	code := warpdrive.OfferAwaiting
	if peer.Mod == warpdrive.PeerModTrust {
		err = d.Download(offerId)
		if err != nil {
			d.log.Println("Cannot auto accept files offer", offerId, err)
		} else {
			code = warpdrive.OfferAccepted
		}
	}
	// Send received
	_ = d.cslq.Encodef("c", code)
	return
}

func (d Dispatcher) Upload(
	offerId warpdrive.OfferId,
	index int,
	offset int64,
) (err error) {
	srv := d.srv.Outgoing()

	// Obtain setup service with offer id
	offer := srv.Get(offerId)
	if offer == nil {
		err = warpdrive.Error(nil, "Cannot find offer with id", offerId)
		return
	}

	// Update status
	srv.Accept(offer)

	// Send confirmation
	err = d.cslq.Encodef("c", 0)
	if err != nil {
		err = warpdrive.Error(err, "Cannot send confirmation")
		return
	}

	finish := make(chan error)

	// Send files
	go func() {
		defer close(finish)
		// Copy files to connection
		offer.Index = index
		offer.Progress = offset
		err := srv.Copy(offer).To(d.conn)
		if err != nil {
			err = warpdrive.Error(err, "Cannot upload files")
		}
		finish <- err
	}()

	// Read OK
	go func() {
		var code byte
		if err := d.cslq.Decodef("c", &code); err != nil {
			err = warpdrive.Error(err, "Cannot read ok")
			d.log.Println(err)
			_ = d.conn.Close()
		}
	}()

	// Ensure the status will be updated
	func() {
		d.srv.Job().Add(1)
		select {
		case err = <-finish:
		case <-d.ctx.Done():
			_ = d.conn.Close()
			err = <-finish
		}
		srv.Finish(offer, err)
		time.Sleep(200)
		d.srv.Job().Done()
	}()
	return
}
