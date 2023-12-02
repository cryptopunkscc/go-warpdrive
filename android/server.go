package android

import (
	"context"
	"github.com/cryptopunkscc/go-apphost-jrpc/android/notify"
	"github.com/cryptopunkscc/go-warpdrive/jrpc"
	"github.com/cryptopunkscc/go-warpdrive/service"
	"github.com/cryptopunkscc/go-warpdrive/storage"
	"log"
)

func Server(ctx context.Context, cache string, store string) error {
	logger := log.New(log.Writer(), "[warpdrive] ", 0)
	factory := storage.NewFactory(logger, cache, store)
	factory.FileResolver = NewResolver()
	createNotify := CreateNotify(notify.NewClient())
	srv := service.Start(ctx, logger, createNotify, factory)
	err := jrpc.Start(ctx, logger, srv)
	<-srv.Done()
	return err
}
