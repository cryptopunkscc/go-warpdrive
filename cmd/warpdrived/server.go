package main

import (
	"go-warpdrive/server"
	"go-warpdrive/service"
	"go-warpdrive/storage/file"
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
