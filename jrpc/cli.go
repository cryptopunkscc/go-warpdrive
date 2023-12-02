package jrpc

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	rpc "github.com/cryptopunkscc/go-apphost-jrpc"
	"github.com/cryptopunkscc/go-warpdrive"
	"io"
)

func Cli(ctx context.Context) error {
	s := rpc.Server[any]{}
	s.Accept = func(query *astral.QueryData) (conn io.ReadWriteCloser, err error) {
		if query.RemoteIdentity() == id.Anyone {
			return query.Accept()
		}
		return nil, errors.New("rejected")
	}
	s.Handler = func(ctx context.Context, conn *rpc.Conn) any {
		if conn == nil {
			return warpdrive.PortCli
		}
		c := warpdrive.Cli{
			Conn:   conn,
			Client: NewClient(),
		}
		return c.Serve(ctx)
	}
	return s.Run(ctx)
}
