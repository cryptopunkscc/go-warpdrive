package proto

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"go-warpdrive/adapter"
	"io"
	"log"
)

type Client struct {
	api       adapter.Api
	conn      io.ReadWriteCloser
	cslq      *cslq.Endec
	log       *log.Logger
	localNode string
}

func NewClient(api adapter.Api) Client {
	return Client{log: log.Default(), api: api}
}

// Connect to warpdrive service
func (c Client) Connect(identity id.Identity, port string) (client Client, err error) {
	c.log = NewLogger("[CLIENT]", port)
	// Resolve local id
	localId, err := c.api.Resolve("localnode")
	if err != nil {
		c.log.Println("Cannot resolve local node id", err)
		return
	}
	c.localNode = localId.String()
	// Connect to local service
	c.conn, err = c.api.Query(identity, port)
	if err != nil {
		c.log.Println("Cannot connect to service", err)
		return
	}
	c.cslq = cslq.NewEndec(c.conn)
	client = c
	return
}

func (c Client) Close() (err error) {
	err = c.cslq.Encode("c", cmdClose)
	_ = c.conn.Close()
	return
}
