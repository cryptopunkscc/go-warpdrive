package main

import (
	"context"
	"github.com/cryptopunkscc/go-warpdrive/proto"
	"github.com/cryptopunkscc/go-warpdrive/service"
	"github.com/cryptopunkscc/go-warpdrive/storage"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// The warpdrive launcher for desktop.
func main() {
	logger := log.New(os.Stderr, "warpdrive ", log.LstdFlags|log.Lmsgprefix)

	// Set up app execution context
	ctx, shutdown := context.WithCancel(context.Background())

	// Trap ctrl+c
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT)
	go func() {
		for {
			<-sigCh
			logger.Println("shutting down...")
			shutdown()

			<-sigCh
			logger.Println("forcing shutdown...")
			os.Exit(0)
		}
	}()

	cache := cacheDir()
	store := storageDir()
	factory := storage.NewFactory(logger, cache, store)
	srv := service.Start(ctx, logger, nil, factory)
	if err := proto.Start(ctx, logger, srv); err != nil {
		logger.Println("cannot run server:", err)
		os.Exit(1)
	}

	<-ctx.Done()
	<-srv.Done()

	time.Sleep(50 * time.Millisecond)
}

func storageDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Panicln("cannot resolve home dir", err)
	}
	dir = filepath.Join(dir, "warpdrive")
	if err = os.MkdirAll(dir, 0700); err != nil {
		log.Panicln("cannot create storage dir", err)
	}
	return dir
}

func cacheDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		log.Panicln("cannot resolve config dir:", err)
	}
	dir = filepath.Join(dir, "warpdrive")
	if err = os.MkdirAll(dir, 0700); err != nil {
		log.Panicln("", err)
	}
	return dir
}
