package embedded

//
//import (
//	"context"
//	"github.com/cryptopunkscc/astrald/auth/id"
//	"github.com/cryptopunkscc/astrald/net"
//	"github.com/cryptopunkscc/astrald/node/modules"
//	"github.com/cryptopunkscc/astrald/node/services"
//	"go-warpdrive/adapter"
//	"io"
//)
//
//var _ adapter.Api = &Adapter{}
//
//type Adapter struct {
//	Ctx  context.Context
//	Node modules.Node
//}
//
//func (a *Adapter) Resolve(name string) (identity id.Identity, err error) {
//	if name == "localnode" {
//		return a.Node.Identity(), nil
//	}
//
//	if identity, err = id.ParsePublicKeyHex(name); err == nil {
//		return
//	}
//
//	identity, err = a.Node.Resolver().Resolve(name)
//	return
//}
//
//type astralPort struct {
//	Node modules.Node
//	Port chan *services.Query
//	*services.Service
//}
//
//type astralRequest struct {
//	Node  modules.Node
//	query *services.Query
//}
//
//func (a *Adapter) Register(name string) (p adapter.Port, err error) {
//	port := astralPort{
//		Node: a.Node,
//		Port: make(chan *services.Query, 128),
//	}
//	p = &port
//	port.Service, err = a.Node.Services().Register(a.Ctx, a.Node.Identity(), name,
//		func(ctx context.Context, query *services.Query) error {
//			port.Port <- query
//			return nil
//		},
//	)
//	return
//}
//
//func (a *Adapter) Query(nodeID id.Identity, query string) (rwc io.ReadWriteCloser, err error) {
//	return a.Node.Query(a.Ctx, nodeID, query)
//}
//
//func (p *astralPort) Next() <-chan adapter.Request {
//	c := make(chan adapter.Request)
//	go func() {
//		defer close(c)
//		for query := range p.Port {
//			q := query
//			c <- &astralRequest{p.Node, q}
//		}
//	}()
//	return c
//}
//
//func (p *astralPort) Close() error {
//	return p.Service.Close()
//}
//
//func (r *astralRequest) Caller() id.Identity {
//	return r.query.RemoteIdentity()
//}
//
//func (r *astralRequest) Query() string {
//	return r.query.Query()
//}
//
//func (r *astralRequest) Accept() (io.ReadWriteCloser, error) {
//	return r.query.Accept()
//}
//
//func (r *astralRequest) Reject() error {
//	return r.query.Reject()
//}
