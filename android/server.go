package android

import (
	"go-warpdrive/adapter"
	"go-warpdrive/server"
	"go-warpdrive/service"
	"path/filepath"
)

func Server(dir string, api adapter.Api) *warpdrived.Server {
	return &warpdrived.Server{
		Component: service.Component{
			Config: service.Config{
				Platform:      service.PlatformAndroid,
				RepositoryDir: filepath.Join(dir, "warpdrive"),
			},
			Sys: &service.Sys{
				Notify: NewNotifier(api),
			},
			FileResolver: NewResolver(api),
		},
	}
}
