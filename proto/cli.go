package proto

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/astral"
	rpc "github.com/cryptopunkscc/go-apphost-jrpc"
	"github.com/cryptopunkscc/go-warpdrive"
	"io"
	"log"
	"strings"
)

const localnode = "localnode"
const prompt = "warp> "

type Cli struct {
	Conn      io.ReadWriteCloser
	Log       *log.Logger
	logPrefix string
}

func Run(ctx context.Context) error {
	s := rpc.Server[any]{}
	s.Ctx = ctx
	s.Accept = func(query *astral.QueryData) (conn *astral.Conn, err error) {
		if query.RemoteIdentity() == id.Anyone {
			return query.Accept()
		}
		return nil, errors.New("rejected")
	}
	s.Handler = func(conn *rpc.Conn) any {
		if conn == nil {
			return warpdrive.PortCli
		}
		c := Cli{Conn: conn}
		return c.Serve(ctx)
	}
	return s.Run()
}

func (c Cli) Serve(ctx context.Context) (err error) {
	localId, err := astral.Resolve(localnode)
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, localnode, localId.String())
	prompt := prompt
	scanner := bufio.NewScanner(c.Conn)
	_, err = c.Conn.Write([]byte(prompt))
	if err != nil {
		err = warpdrive.Error(err, "Cannot write prompt")
		return
	}
	conn, err := NewClient().Connect(id.Identity{}, warpdrive.Port)
	if err != nil {
		err = warpdrive.Error(err, "Cannot connect local client")
		return
	}
	finish := make(chan struct{})
	defer close(finish)
	go func() {
		select {
		case <-ctx.Done():
		case <-finish:
		}
		_ = c.Conn.Close()
		_ = conn.Close()
	}()
	_cslq := cslq.NewEndec(c.Conn)
	for scanner.Scan() {
		text := scanner.Text()

		switch text {
		case "prompt-off":
			prompt = ""
			_ = _cslq.Encodef("[c]c", "\n")
			continue
		case "e", "exit":
			return
		case "", "h", "help":
			_ = cliHelp(ctx, c.Conn, conn, nil)
			_ = _cslq.Encodef("[c]c", prompt)
			continue
		}

		words := strings.Split(text, " ")
		if len(words) == 0 {
			_ = _cslq.Encodef("[c]c", prompt)
			continue
		}
		cmd, args := words[0], words[1:]
		//c.Log = warpdrive.NewLogger(c.logPrefix, fmt.Sprintf("(%s)", cmd))
		c.Log = log.Default()
		fn, ok := commands[cmd]
		if ok {
			err = fn(ctx, c.Conn, conn, args)
			if err != nil {
				c.Log.Println(err)
			}
			//d.Println("OK")
		} else {
			c.Log.Println("no such cli command", cmd)
		}
		_ = _cslq.Encodef("[c]c", prompt)
	}
	return scanner.Err()
}

var commands = cmdMap{
	"peers":  cliPeers,
	"send":   cliSend,
	"out":    cliSent,
	"in":     cliReceived,
	"sub":    cliSubscribe,
	"get":    cliDownload,
	"update": cliUpdate,
	"stat":   cliStatus,
}

type cmdMap map[string]cmdFunc
type cmdFunc func(context.Context, io.ReadWriteCloser, warpdrive.Client, []string) error

var filters = map[string]warpdrive.Filter{
	"all": warpdrive.FilterAll,
	"in":  warpdrive.FilterIn,
	"out": warpdrive.FilterOut,
}

// =========================== Commands ===============================

func cliHelp(ctx context.Context, writer io.ReadWriteCloser, _ warpdrive.Client, _ []string) (err error) {
	for name := range commands {
		if _, err = fmt.Fprintln(writer, name); err != nil {
			return err
		}
	}
	return
}

func cliPeers(ctx context.Context, writer io.ReadWriteCloser, client warpdrive.Client, _ []string) (err error) {
	peers, err := client.ListPeers()
	if err != nil {
		return
	}
	for _, peer := range peers {
		_, err = fmt.Fprintln(writer, peer.Id, peer.Alias, peer.Mod)
		if err != nil {
			return
		}
	}
	return
}

func cliSend(ctx context.Context, writer io.ReadWriteCloser, client warpdrive.Client, args []string) (err error) {
	if len(args) < 1 {
		_, err = fmt.Fprintln(writer, "<filePath> <peerId>?")
		return
	}

	peer := ctx.Value(localnode).(string)
	if len(args) > 1 {
		peer = args[1]
	}
	offer, err := client.CreateOffer(warpdrive.PeerId(peer), args[0])
	if err != nil {
		return err
	}
	status := "delivered"
	if offer.Status == warpdrive.StatusAccepted {
		status = "accepted"
	}
	_, err = fmt.Fprintln(writer, offer.Id, status)
	return
}

func cliSent(ctx context.Context, writer io.ReadWriteCloser, client warpdrive.Client, _ []string) (err error) {
	sent, err := client.ListOffers(warpdrive.FilterOut)
	if err != nil {
		return err
	}
	for _, offer := range sent {
		err = printFilesRequest(writer, offer)
		if err != nil {
			return
		}
	}
	return
}

func cliReceived(ctx context.Context, writer io.ReadWriteCloser, client warpdrive.Client, _ []string) (err error) {
	received, err := client.ListOffers(warpdrive.FilterIn)
	if err != nil {
		return err
	}
	for _, offer := range received {
		err = printFilesRequest(writer, offer)
		if err != nil {
			return
		}
	}
	return
}

func cliSubscribe(ctx context.Context, conn io.ReadWriteCloser, client warpdrive.Client, args []string) (err error) {
	filter := "all"
	if len(args) > 0 {
		filter = args[0]
	}
	f, exist := filters[filter]
	if !exist {
		_, err = fmt.Fprintln(conn, "Invalid filter: ", filter)
		return
	}
	var offers <-chan warpdrive.Offer
	offers, err = client.ListenOffers(f)
	if err != nil {
		return err
	}
	finish := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
		case <-finish:
		}
		_ = client.Close()
		_ = conn.Close()
	}()
	go func() {
		defer close(finish)
		var code byte
		err = cslq.Decode(conn, "c", &code)
		if errors.Is(err, io.EOF) {
			err = nil
		}
	}()
	for offer := range offers {
		_ = printFilesRequest(conn, offer)
	}
	return
}

func cliStatus(ctx context.Context, conn io.ReadWriteCloser, client warpdrive.Client, args []string) (err error) {
	filter := "all"
	if len(args) > 0 {
		filter = args[0]
	}
	f, exist := filters[filter]
	if !exist {
		_, err = fmt.Fprintln(conn, "Invalid filter: ", filter)
		return
	}
	var events <-chan warpdrive.OfferStatus
	events, err = client.ListenStatus(f)
	finish := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
		case <-finish:
		}
		_ = client.Close()
		_ = conn.Close()
	}()
	go func() {
		defer close(finish)
		var code byte
		err = cslq.Decode(conn, "c", &code)
		if errors.Is(err, io.EOF) {
			err = nil
		}
	}()
	for event := range events {
		_, _ = fmt.Fprintln(conn, event.Id, event.Update, event.In, event.Status, event.Index, event.Progress)
	}
	return
}

func cliDownload(ctx context.Context, writer io.ReadWriteCloser, client warpdrive.Client, args []string) (err error) {
	if len(args) < 1 {
		_, err = fmt.Fprintln(writer, "<offerId>")
		return
	}
	err = client.AcceptOffer(warpdrive.OfferId(args[0]))
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(writer, "accepted")
	return
}

func cliUpdate(ctx context.Context, writer io.ReadWriteCloser, client warpdrive.Client, args []string) (err error) {
	if len(args) < 3 {
		_, err = fmt.Fprintln(writer, "<peerId> <attr> <value>")
		return
	}
	err = client.UpdatePeer(warpdrive.PeerId(args[0]), args[1], args[2])
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(writer, "updated")
	return
}

func printFilesRequest(writer io.Writer, offer warpdrive.Offer) (err error) {
	_, err = fmt.Fprintln(writer, "incoming:", offer.In)
	_, err = fmt.Fprintln(writer, "peer:", offer.Peer)
	_, err = fmt.Fprintln(writer, "offer id:", offer.Id)
	_, err = fmt.Fprintln(writer, "created at:", offer.Create)
	_, err = fmt.Fprintln(writer, "status:", offer.Status)
	if offer.Index > -1 {
		_, err = fmt.Fprintln(writer, "  file index:", offer.Index)
		_, err = fmt.Fprintln(writer, "  progress:", offer.Progress)
		_, err = fmt.Fprintln(writer, "  update at:", offer.Update)
	}
	_, err = fmt.Fprintln(writer, "files:")
	if err != nil {
		return
	}
	for i, file := range offer.Files {
		_, err = fmt.Fprintf(writer, "%d. %s (%dB)\n", i, file.Name, file.Size)
		if err != nil {
			return
		}
	}
	_, err = fmt.Fprintln(writer, "-----------------------------------")
	return
}
