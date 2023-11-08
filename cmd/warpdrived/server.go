package main

import (
	"github.com/cryptopunkscc/go-warpdrive/server"
	"github.com/cryptopunkscc/go-warpdrive/service"
	"github.com/cryptopunkscc/go-warpdrive/storage/file"
)

func Server() *warpdrived.Server {
	return &warpdrived.Server{
		Component: service.Component{
			Config: service.Config{
				Platform: service.PlatformDesktop,
			},
			FileResolver: file.Resolver{},
		},
	}
}
