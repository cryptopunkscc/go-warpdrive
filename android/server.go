package android

import (
	"github.com/cryptopunkscc/go-warpdrive/proto"
	"github.com/cryptopunkscc/go-warpdrive/service"
	"path/filepath"
)

func Server(dir string) *proto.Server {
	srv := service.NewComponent()
	srv.Config = service.Config{
		Platform:      service.PlatformAndroid,
		RepositoryDir: filepath.Join(dir, "warpdrive"),
	}
	srv.Sys = &service.Sys{
		Notify: NewNotifier(),
	}
	srv.FileResolver = NewResolver()
	return &proto.Server{Service: &srv}
}
