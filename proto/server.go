package proto

import (
	"context"
	rpc "github.com/cryptopunkscc/go-apphost-jrpc"
	"github.com/cryptopunkscc/go-warpdrive"
	"log"
)

func Start(
	ctx context.Context,
	logger *log.Logger,
	srv warpdrive.Service,
) (err error) {
	l := logger

	c := Client{}
	c.Ctx = ctx

	s := rpc.Server[warpdrive.Api]{
		Ctx: ctx,

		Handler: func(conn *rpc.Conn) (h warpdrive.Api) {
			if conn != nil {
				conn.WithLogger(l)
				h = warpdrive.NewHandler(
					conn.Conn,
					ctx,
					&c,
					srv,
					l,
				)
			} else {
				h = warpdrive.Handler{}
			}
			return
		},
	}

	return s.Run()
}
