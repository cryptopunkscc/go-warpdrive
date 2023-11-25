package proto

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	rpc "github.com/cryptopunkscc/go-apphost-jrpc"
	"github.com/cryptopunkscc/go-warpdrive"
)

type Client struct {
	rpc.Conn
}

func (c Client) Notify() error {
	return c.Encode(nil)
}

func (c Client) Await() (err error) {
	return rpc.Await(c.Conn)
}

func NewClient() warpdrive.Client {
	return &Client{}
}

func (c Client) Connect(identity id.Identity, port string) (client warpdrive.Client, err error) {
	conn, err := astral.Query(identity, port)
	if err != nil {
		return
	}
	client = c.Attach(*conn)
	return
}

func (c Client) Attach(conn astral.Conn) (client warpdrive.Client) {
	if c.Ctx == nil {
		c.Ctx = context.Background()
	}

	c.Conn = *rpc.NewConn(c.Ctx, &conn)
	client = &c
	return
}

func (c Client) CreateOffer(peerId warpdrive.PeerId, filePath string) (status warpdrive.OfferStatus, err error) {
	return rpc.Query[warpdrive.OfferStatus](c.Conn, "CreateOffer", peerId, filePath)
}

func (c Client) AcceptOffer(id warpdrive.OfferId) (err error) {
	return rpc.Command(c.Conn, "AcceptOffer", id)
}

func (c Client) ListOffers(filter warpdrive.Filter) (offers []warpdrive.Offer, err error) {
	return rpc.Query[[]warpdrive.Offer](c.Conn, "ListOffers", filter)
}

func (c Client) ListenStatus(filter warpdrive.Filter) (status <-chan warpdrive.OfferStatus, err error) {
	return rpc.Subscribe[warpdrive.OfferStatus](c.Conn, "ListenStatus", filter)
}

func (c Client) ListenOffers(filter warpdrive.Filter) (out <-chan warpdrive.Offer, err error) {
	return rpc.Subscribe[warpdrive.Offer](c.Conn, "ListenOffers", filter)
}

func (c Client) ListPeers() (peers []warpdrive.Peer, err error) {
	return rpc.Query[[]warpdrive.Peer](c.Conn, "ListPeers")
}

func (c Client) UpdatePeer(peerId warpdrive.PeerId, attr string, value string) (err error) {
	return rpc.Command(c.Conn, "UpdatePeer", attr, value)
}

func (c Client) SendOffer(offerId warpdrive.OfferId, files []warpdrive.Info) (accepted bool, err error) {
	return rpc.Query[bool](c.Conn, "SendOffer", offerId, files)
}

func (c Client) Download(offerId warpdrive.OfferId, index int, offset int64) (err error) {
	return rpc.Command(c.Conn, "Download", offerId, index, offset)
}
