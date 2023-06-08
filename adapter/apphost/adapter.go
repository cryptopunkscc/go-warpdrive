package apphost

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	apphost "github.com/cryptopunkscc/astrald/lib/astral"
	"go-warpdrive/adapter"
	"io"
)

var _ adapter.Api = Adapter{}

type Adapter struct{}

func (a Adapter) Resolve(name string) (id.Identity, error) {
	return apphost.Resolve(name)
}

func (a Adapter) Register(name string) (adapter.Port, error) {
	listener, err := apphost.Register(name)
	if err != nil {
		return nil, err
	}
	return appHostPort{listener}, err
}

func (a Adapter) Query(nodeID id.Identity, query string) (rw io.ReadWriteCloser, err error) {
	return apphost.Query(nodeID, query)
}

type appHostPort struct{ *apphost.Listener }

func (a appHostPort) Next() <-chan adapter.Request {
	c := make(chan adapter.Request)
	go func() {
		defer close(c)
		for query := range a.QueryCh() {
			q := query
			c <- &appHostRequest{q}
		}
	}()
	return c
}

func (a appHostPort) Close() error {
	return a.Listener.Close()
}

type appHostRequest struct{ query *apphost.QueryData }

func (a appHostRequest) Caller() id.Identity {
	return a.query.RemoteIdentity()
}

func (a appHostRequest) Accept() (io.ReadWriteCloser, error) {
	return a.query.Accept()
}

func (a appHostRequest) Reject() error {
	return a.query.Reject()
}

func (a appHostRequest) Query() string {
	return a.query.Query()
}
