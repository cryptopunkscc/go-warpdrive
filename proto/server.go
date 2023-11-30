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
	s := rpc.Server[warpdrive.Api]{
		Handler: func(ctx2 context.Context, conn *rpc.Conn) (h warpdrive.Api) {
			if conn != nil {
				conn.WithLogger(l)
				h = warpdrive.NewHandler(
					ctx,
					ctx2,
					conn,
					&Client{},
					srv,
					l,
					conn.RemoteIdentity(),
				)
			} else {
				h = warpdrive.Handler{}
			}
			return
		},
	}

	return s.Run(ctx)
}
