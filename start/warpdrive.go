package start

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/go-warpdrive"
	"github.com/cryptopunkscc/go-warpdrive/jrpc"
	"github.com/cryptopunkscc/go-warpdrive/service"
	"github.com/cryptopunkscc/go-warpdrive/storage"
	"io"
	"log"
)

type Args struct {
	Logger       *log.Logger
	Cache        string
	Store        string
	CreateNotify warpdrive.CreateNotify
}

func Warpdrive(
	ctx context.Context,
	args Args,
) error {
	if args.Logger == nil {
		args.Logger = log.New(io.Discard, "", 0)
	}
	factory := storage.NewFactory(args.Logger, args.Cache, args.Store)
	srv := service.Start(ctx, args.Logger, args.CreateNotify, factory)
	if err := jrpc.Start(ctx, args.Logger, srv); err != nil {
		return fmt.Errorf("cannot run server: %v", err)
	}

	<-ctx.Done()
	<-srv.Done()
	return nil
}
