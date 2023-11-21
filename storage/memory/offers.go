package memory

import (
	"github.com/cryptopunkscc/go-warpdrive"
)

type Offer warpdrive.Offers

var _ warpdrive.OfferStorage = Offer{}

func (r Offer) Save(offer warpdrive.Offer) {
	r[offer.Id] = &offer
}

func (r Offer) Get() warpdrive.Offers {
	return warpdrive.Offers(r)
}
