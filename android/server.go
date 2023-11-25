package android

import (
	"context"
	"github.com/cryptopunkscc/go-warpdrive/proto"
	"github.com/cryptopunkscc/go-warpdrive/service"
	"github.com/cryptopunkscc/go-warpdrive/storage"
	"log"
)

func Server(ctx context.Context, cache string, store string) error {
	logger := log.New(log.Writer(), "[warpdrive] ", 0)
	factory := storage.NewFactory(logger, cache, store)
	factory.FileResolver = NewResolver()
	notify := NewNotifier()
	srv := service.Start(ctx, logger, notify, factory)
	return proto.Start(ctx, logger, srv)
}
