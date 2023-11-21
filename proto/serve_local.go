package proto

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/go-warpdrive"
	"io"
	"log"
	"time"
)

func (d Dispatcher) CreateOffer(peerId warpdrive.PeerId, filePath string) (err error) {
	// Get files info
	files, err := d.srv.File().Info(filePath)
	if err != nil {
		err = warpdrive.Error(err, "Cannot get files info")
		return
	}

	// Parse identity
	identity, err := id.ParsePublicKeyHex(string(peerId))
	if err != nil {
		err = warpdrive.Error(err, "Cannot parse peer id")
		return
	}

	// Connect to remote client
	client, err := NewClient().Connect(identity, warpdrive.Port)
	if err != nil {
		err = warpdrive.Error(err, "Cannot connect to remote", peerId)
		return
	}

	// Send file to recipient service
	offerId := warpdrive.NewOfferId()
	code, err := client.SendOffer(offerId, files)
	_ = client.Close()
	if err != nil {
		err = warpdrive.Error(err, "Cannot send file")
		return
	}

	d.srv.Outgoing().Add(offerId, files, peerId)

	// Write id to sender
	err = d.cslq.Encodef("[c]c c", offerId, code)
	if err != nil {
		err = warpdrive.Error(err, "Cannot send create offer result", offerId)
		return
	}
	d.log.Println(filePath, "offer sent to", peerId)
	return
}

func (d Dispatcher) ListOffers(filter warpdrive.Filter) (err error) {
	// Collect file offers
	offers := warpdrive.FilterOffers(d.srv, filter)
	d.log.Println("Filter", filter)
	// Send filtered file offers
	if err = json.NewEncoder(d.conn).Encode(offers); err != nil {
		err = warpdrive.Error(err, "Cannot send incoming offers")
		return
	}
	return
}

func (d Dispatcher) AcceptOffer(offerId warpdrive.OfferId) (err error) {
	// Download offer
	d.log.Println("Accepted incoming files", offerId)
	err = d.Download(offerId)
	if err != nil {
		err = warpdrive.Error(err, "Cannot download incoming files", offerId)
		return
	}
	// Send ok
	err = d.cslq.Encodef("c", 0)
	if err != nil {
		err = warpdrive.Error(err, "Cannot send ok")
		return
	}
	return
}

func (d Dispatcher) Download(offerId warpdrive.OfferId) (err error) {
	// Get incoming offer service for offer id
	srv := d.srv.Incoming()
	offer := srv.Get(offerId)
	if offer == nil {
		err = warpdrive.Error(nil, "Cannot find incoming file")
		return
	}

	// parse peer id
	peerId, err := id.ParsePublicKeyHex(string(offer.Peer))
	if err != nil {
		err = warpdrive.Error(err, "Cannot parse peer id", offer.Peer)
		return
	}

	// Update status
	srv.Accept(offer)

	// Connect to remote warpdrive
	client, err := NewClient().Connect(peerId, warpdrive.Port)
	if err != nil {
		return
	}

	// Request download
	if err = client.Download(offerId, offer.Index, offer.Progress); err != nil {
		err = warpdrive.Error(err, "Cannot download offer")
		return err
	}

	finish := make(chan error)

	// Ensure the status will be updated
	go func() {
		d.srv.Job().Add(1)
		select {
		case err = <-finish:
		case <-d.ctx.Done():
			_ = client.conn.Close()
			err = <-finish
		}
		if err != nil {
			d.log.Println(warpdrive.Error(err, "Failed"))
		}
		_ = client.Close()
		srv.Finish(offer, err)
		time.Sleep(200)
		d.srv.Job().Done()
	}()

	// download files in background
	go func() {
		defer close(finish)
		// Copy files from connection to storage
		if err = srv.Copy(offer).From(client.conn); err != nil {
			finish <- warpdrive.Error(err, "Cannot download files")
			return
		}
		// Send OK
		if err = client.cslq.Encodef("c", 0); err != nil {
			finish <- warpdrive.Error(err, "Cannot send ok")
			return
		}
		finish <- nil
	}()
	return
}

func (d Dispatcher) ListPeers(any) (err error) {
	// Get peers
	peers := d.srv.Peer().List()
	// Send peers
	if err = json.NewEncoder(d.conn).Encode(peers); err != nil {
		err = warpdrive.Error(err, "Cannot send peers")
		return
	}
	return
}

func (d Dispatcher) ListenStatus(filter warpdrive.Filter) (err error) {
	unsub := d.filterSubscribe(filter, warpdrive.OfferService.StatusSubscriptions)
	defer unsub()
	// Wait for close
	var code byte
	err = d.cslq.Decodef("c", &code)
	return
}

func (d Dispatcher) ListenOffers(filter warpdrive.Filter) (err error) {
	unsub := d.filterSubscribe(filter, warpdrive.OfferService.OfferSubscriptions)
	defer unsub()
	// Wait for close
	var code byte
	err = d.cslq.Decodef("c", &code)
	return
}
func (d Dispatcher) filterSubscribe(
	filter warpdrive.Filter,
	get func(service warpdrive.OfferService) *warpdrive.Subscriptions,
) (unsub warpdrive.Unsubscribe) {
	c := NewListener(d.ctx, d.conn)
	var unsubIn warpdrive.Unsubscribe = func() {}
	var unsubOut warpdrive.Unsubscribe = func() {}
	switch filter {
	case warpdrive.FilterIn:
		unsubIn = get(d.srv.Incoming()).Subscribe(c)
	case warpdrive.FilterOut:
		unsubOut = get(d.srv.Outgoing()).Subscribe(c)
	default:
		unsubIn = get(d.srv.Incoming()).Subscribe(c)
		unsubOut = get(d.srv.Outgoing()).Subscribe(c)
	}
	return func() {
		unsubIn()
		unsubOut()
		close(c)
	}
}
func NewListener(ctx context.Context, w io.WriteCloser) (listener warpdrive.Listener) {
	c := make(chan interface{}, 1024)
	listener = c
	e := json.NewEncoder(w)
	go func() {
		for {
			select {
			case <-ctx.Done():
				_ = w.Close()
				return
			case i, ok := <-c:
				if !ok {
					return
				}
				var err error
				switch v := i.(type) {
				case []byte:
					v = append(v, '\n')
					_, err = w.Write(v)
				default:
					err = e.Encode(i)
				}
				if err != nil {
					log.Println("Cannot write", err)
					return
				}
			}
		}
	}()
	return
}

func (d Dispatcher) UpdatePeer(any) (err error) {
	// Read peer update
	// Fixme refactor to cslq
	var req []string
	if err = json.NewDecoder(d.conn).Decode(&req); err != nil {
		err = warpdrive.Error(err, "Cannot read peer update")
		return
	}
	peerId := req[0]
	attr := req[1]
	value := req[2]
	// Update peer
	d.srv.Peer().Update(warpdrive.PeerId(peerId), attr, value)
	// Send OK
	if err = d.cslq.Encodef("c", 0); err != nil {
		err = warpdrive.Error(err, "Cannot send ok")
		return
	}
	return
}
