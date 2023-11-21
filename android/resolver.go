package android

import (
	"github.com/cryptopunkscc/go-apphost-jrpc/android"
	"github.com/cryptopunkscc/go-apphost-jrpc/android/content"
	"github.com/cryptopunkscc/go-warpdrive"
	"path"
)

type resolver struct {
	content.Client
}

func NewResolver() warpdrive.FileResolver {
	return &resolver{}
}

func (c resolver) Info(uri string) (files []warpdrive.Info, err error) {
	i, err := c.Client.Info(uri)
	if err != nil {
		return
	}
	files = append(files, warpdrive.Info{
		Uri:   i.Uri,
		Size:  i.Size,
		Mime:  i.Mime,
		Name:  resolveName(i),
		IsDir: false,
		Perm:  0755,
	})
	return
}

func resolveName(i android.Info) string {
	switch {
	case i.Name != "":
		return i.Name
	default:
		return path.Base(i.Uri)
	}
}
