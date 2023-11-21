package warpdrive

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type Handler struct {
	io.ReadWriteCloser
	ctx      context.Context
	parent   context.Context
	client   Client
	srv      Service
	logger   *log.Logger
	remoteID id.Identity
}

func NewHandler(
	parent context.Context,
	ctx context.Context,
	conn io.ReadWriteCloser,
	client Client,
	srv Service,
	log *log.Logger,
	remoteID id.Identity,
) Api {
	return &Handler{
		ReadWriteCloser: conn,
		ctx:             ctx,
		parent:          parent,
		client:          client,
		srv:             srv,
		logger:          log,
		remoteID:        remoteID,
	}
}

func (d Handler) String() string {
	return Port
}

// ============================ local ============================

func (d Handler) CreateOffer(peerId PeerId, filePath string) (os OfferStatus, err error) {
	// Get files info
	files, err := d.srv.File().Info(filePath)
	if err != nil {
		err = Error(err, "Cannot get files info")
		return
	}

	// Parse identity
	identity := id.Identity{}
	if peerId != "" {
		identity, err = id.ParsePublicKeyHex(string(peerId))
		if err != nil {
			err = Error(err, "Cannot parse peer id")
			return
		}
	}

	// Connect to remote client
	client, err := d.client.Connect(identity, Port)
	if err != nil {
		err = Error(err, "Cannot connect to remote", peerId)
		return
	}

	// Send file to recipient service
	os.Id = NewOfferId()
	shrunken := shrinkPaths(files)
	accepted, err := client.SendOffer(os.Id, shrunken)
	_ = client.Close()
	if err != nil {
		err = Error(err, "Cannot send file")
		return
	}
	d.srv.Outgoing().Add(os.Id, files, peerId)

	// Setup result
	os.Status = StatusAwaiting
	if accepted {
		os.Status = StatusAccepted
	}
	return
}

// TODO make it bulletproof
func shrinkPaths(in []Info) (out []Info) {
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

func (d Handler) AcceptOffer(offerId OfferId) (err error) {
	// Download offer
	d.logger.Println("Accepted incoming files", offerId)
	if err = d.downloadAsync(offerId); err != nil {
		err = Error(err, "Cannot download incoming files", offerId)
		return
	}
	return
}

func (d Handler) ListOffers(filter Filter) (offers []Offer, err error) {
	// Collect file offers
	offers = FilterOffers(d.srv, filter)
	return
}

func (d Handler) ListenStatus(filter Filter) (out <-chan OfferStatus, err error) {
	s := make(chan OfferStatus)
	out = s
	switch filter {
	case FilterIn:
		d.srv.Incoming().StatusBroadcast().Listen(d.ctx, s)
	case FilterOut:
		d.srv.Outgoing().StatusBroadcast().Listen(d.ctx, s)
	default:
		d.srv.Outgoing().StatusBroadcast().Listen(d.ctx, s)
		d.srv.Incoming().StatusBroadcast().Listen(d.ctx, s)
	}
	return
}

func (d Handler) ListenOffers(filter Filter) (out <-chan Offer, err error) {
	s := make(chan Offer)
	out = s
	switch filter {
	case FilterIn:
		d.srv.Incoming().OfferBroadcast().Listen(d.ctx, s)
	case FilterOut:
		d.srv.Outgoing().OfferBroadcast().Listen(d.ctx, s)
	default:
		d.srv.Outgoing().OfferBroadcast().Listen(d.ctx, s)
		d.srv.Incoming().OfferBroadcast().Listen(d.ctx, s)
	}
	return
}

func (d Handler) ListPeers() (peers []Peer, err error) {
	peers = d.srv.Peer().List()
	return
}

func (d Handler) UpdatePeer(peerId PeerId, attr string, value string) (err error) {
	// Update peer
	d.srv.Peer().Update(peerId, attr, value)
	return
}

func (d Handler) downloadAsync(offerId OfferId) (err error) {
	// Get incoming offer service for offer id
	srv := d.srv.Incoming()
	offer := srv.Get(offerId)
	if offer == nil {
		err = Error(nil, "Cannot find incoming file")
		return
	}

	// parse peer id
	peerId, err := id.ParsePublicKeyHex(string(offer.Peer))
	if err != nil {
		err = Error(err, "Cannot parse peer id", offer.Peer)
		return
	}

	// Update status
	srv.Accept(offer)

	// Connect to remote warpdrive
	client, err := d.client.Connect(peerId, Port)
	if err != nil {
		return
	}

	// Request download
	if err = client.Download(offerId, offer.Index, offer.Progress); err != nil {
		err = Error(err, "Cannot download offer")
		_ = client.Close()
		return err
	}
	ctx, cancelFunc := context.WithCancelCause(context.Background())
	// Ensure connection closed
	go func() {
		select {
		case <-d.parent.Done():
		case <-ctx.Done():
		}
		<-ctx.Done()
		_ = client.Close()
		if err = ctx.Err(); errors.Is(err, context.Canceled) {
			err = nil
		}
		srv.Finish(offer, err)
		time.Sleep(200)
		d.srv.Job().Done()
	}()

	// Download in background
	go func() {
		defer cancelFunc(err)
		d.srv.Job().Add(1)

		if err = client.Notify(); err != nil {
			err = Error(err, "Cannot download files")
			return
		}

		if err = srv.Copy(offer).From(client); err != nil {
			err = Error(err, "Cannot download files")
			return
		}
	}()
	return
}

// ============================ remote ============================

func (d Handler) SendOffer(offerId OfferId, files []Info) (accepted bool, err error) {
	peerId := PeerId(d.remoteID.String())
	peer := d.srv.Peer().Get(peerId)
	// Check if peer is blocked
	if peer.Mod == PeerModBlock {
		d.logger.Println("Blocked request from", peerId)
		return
	}

	// Store incoming offer
	d.srv.Incoming().Add(offerId, files, peerId)
	// Auto accept offer if peer is trusted
	//code := warpdrive.OfferAwaiting
	accepted = false
	if peer.Mod == PeerModTrust {
		err = d.downloadAsync(offerId)
		if err != nil {
			d.logger.Println("Cannot auto accept files offer", offerId, err)
		} else {
			accepted = true
			//code = warpdrive.OfferAccepted
		}
	}
	return
}

func (d Handler) Download(offerId OfferId, index int, offset int64) (err error) {
	srv := d.srv.Outgoing()

	// Obtain setup service with offer id
	var offer *Offer
	if offer = srv.Get(offerId); offer == nil {
		return Error(nil, "Cannot find offer with id", offerId)
	}

	// Update status
	srv.Accept(offer)

	c := d.client.Attach(d)
	if err = c.Notify(); err != nil {
		return
	}
	if err = c.Await(); err != nil {
		return
	}

	d.srv.Job().Add(1)
	offer.Index = index
	offer.Progress = offset
	if err = srv.Copy(offer).To(d); err != nil {
		return Error(err, "Cannot upload files")
	}
	srv.Finish(offer, err)
	time.Sleep(200)
	d.srv.Job().Done()
	return
}
