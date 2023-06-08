package memory

import (
	"go-warpdrive/proto"
	"go-warpdrive/storage"
)

type Offer proto.Offers

var _ storage.Offer = Offer{}

func (r Offer) Save(offer proto.Offer) {
	r[offer.Id] = &offer
}

func (r Offer) Get() proto.Offers {
	return proto.Offers(r)
}
