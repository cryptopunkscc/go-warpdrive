package proto

import (
	uuid "github.com/nu7hatch/gouuid"
	"log"
	"net/url"
	"path/filepath"
	"strings"
)

func (d Dispatcher) filterOffers(
	filter Filter,
) (offers []Offer) {
	switch filter {
	case FilterIn:
		offers = append(offers, d.srv.Incoming().List()...)
	case FilterOut:
		offers = append(offers, d.srv.Outgoing().List()...)
	case FilterAll:
		offers = append(offers, d.srv.Incoming().List()...)
		offers = append(offers, d.srv.Outgoing().List()...)
	}
	return
}

func (d Dispatcher) filterSubscribe(
	filter Filter,
	get func(service OfferService) *Subscriptions,
) (unsub Unsubscribe) {
	c := NewListener(d.ctx, d.conn)
	var unsubIn Unsubscribe = func() {}
	var unsubOut Unsubscribe = func() {}
	switch filter {
	case FilterIn:
		unsubIn = get(d.srv.Incoming()).Subscribe(c)
	case FilterOut:
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

func newOfferId() OfferId {
	v4, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return OfferId(v4.String())
}