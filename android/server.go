package android

import (
	"github.com/cryptopunkscc/go-warpdrive/adapter"
	"github.com/cryptopunkscc/go-warpdrive/server"
	"github.com/cryptopunkscc/go-warpdrive/service"
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
