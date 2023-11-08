package android

import (
	"github.com/cryptopunkscc/go-warpdrive/adapter"
	"github.com/cryptopunkscc/go-warpdrive/android/content"
	warpdrive "github.com/cryptopunkscc/go-warpdrive/proto"
	"github.com/cryptopunkscc/go-warpdrive/storage"
	"path"
)

type resolver struct {
	content.Client
}

func NewResolver(api adapter.Api) storage.FileResolver {
	r := &resolver{}
	r.Api = api
	return r
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

func resolveName(i content.Info) string {
	switch {
	case i.Name != "":
		return i.Name
	default:
		return path.Base(i.Uri)
	}
}
