package warpdrive

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	uuid "github.com/nu7hatch/gouuid"
	"net"
	"os"
)

type Client interface {
	Api
	Connect(identity id.Identity, port string) (client Client, err error)
	Close() (err error)
}

type Api interface {
	net.Conn
	LocalApi
	RemoteApi
}

type LocalApi interface {
	CreateOffer(peerId PeerId, filePath string) (id OfferStatus, err error)
	AcceptOffer(id OfferId) (err error)
	ListOffers(filter Filter) (offers []Offer, err error)
	ListenStatus(filter Filter) (status <-chan OfferStatus, err error)
	ListenOffers(filter Filter) (out <-chan Offer, err error)
	ListPeers() (peers []Peer, err error)
	UpdatePeer(peerId PeerId, attr string, value string) (err error)
}

type RemoteApi interface {
	SendOffer(offerId OfferId, files []Info) (accepted bool, err error)
	Download(offerId OfferId, index int, offset int64) (err error)
}

type Offers map[OfferId]*Offer
type OfferId string
type Offer struct {
	OfferStatus
	// Create time
	Create int64
	// Peer unique identifier
	Peer PeerId
	// Files info
	Files []Info
}

func NewOfferId() OfferId {
	v4, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return OfferId(v4.String())
}

type OfferStatus struct {
	// Id the unique offer identifier.
	Id OfferId
	// In marks if offer is incoming or outgoing.
	In bool
	// Status of the offer
	Status string
	// Index of transferred files. If transfer is not started the index is equal -1.
	Index int
	// Progress of specific file transfer
	Progress int64
	// Update timestamp in milliseconds
	Update int64
}

const (
	OfferAwaiting = 0
	OfferAccepted = 1
)

type Filter uint8

const (
	FilterAll = Filter(iota)
	FilterIn
	FilterOut
)

func (o Offer) IsOngoing() bool {
	return o.Status == StatusUpdated
}

type Peers map[PeerId]*Peer

type PeerId string
type Peer struct {
	Id    PeerId
	Alias string
	Mod   string
}
type Info struct {
	Uri   string
	Path  string
	Size  int64
	IsDir bool
	Perm  os.FileMode
	Mime  string
	Name  string
}
