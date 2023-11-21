package main

import (
	"github.com/cryptopunkscc/go-warpdrive/proto"
	"github.com/cryptopunkscc/go-warpdrive/service"
)

func Server() *proto.Server {
	srv := service.NewComponent()
	srv.Config = service.Config{
		Platform: service.PlatformDesktop,
	}
	return &proto.Server{Service: &srv}
}
