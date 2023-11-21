package proto

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/go-warpdrive"
	"io"
	"log"
)

type Dispatcher struct {
	logPrefix  string
	callerId   string
	authorized bool

	ctx  context.Context
	conn io.ReadWriteCloser

	srv warpdrive.Service

	cslq *cslq.Endec
	log  *log.Logger
}

func NewDispatcher(
	logPrefix string,
	callerId string,
	authorized bool,
	ctx context.Context,
	conn io.ReadWriteCloser,
	srv warpdrive.Service,
) *Dispatcher {
	return &Dispatcher{
		logPrefix:  logPrefix,
		callerId:   callerId,
		authorized: authorized,
		ctx:        ctx,
		conn:       conn,
		srv:        srv,
		cslq:       cslq.NewEndec(conn),
		log:        warpdrive.NewLogger(logPrefix),
	}
}

func (d Dispatcher) Serve(
	dispatch func(d *Dispatcher) error,
) (err error) {
	for err == nil {
		err = dispatch(&d)
		if err == nil {
			d.log.Println("OK")
		}
	}
	if errors.Is(err, warpdrive.ErrEnded) {
		d.log.Println("End")
		err = nil
	}
	if err != nil {
		d.log.Println(warpdrive.Error(err, "Failed"))
	}
	return errors.Unwrap(err)
}

func Dispatch(d *Dispatcher) (err error) {
	cmd, err := nextCommand(d)
	if err != nil {
		return
	}
	switch cmd {
	case infoPing:
		return cslq.Invokef(d.conn, "", d.Ping)
	case remoteSend:
		return cslq.Invokef(d.conn, "", d.Receive)
	case remoteDownload:
		return Invokef3(d.conn, "[c]c q q", d.Upload)
	}
	if !d.authorized {
		return nil
	}
	switch cmd {
	case localListPeers:
		return cslq.Invokef(d.conn, "", d.ListPeers)
	case localCreateOffer:
		return Invokef2(d.conn, "[c]c [c]c", d.CreateOffer)
	case localAcceptOffer:
		return cslq.Invokef(d.conn, "[c]c", d.AcceptOffer)
	case localListOffers:
		return cslq.Invokef(d.conn, "c", d.ListOffers)
	case localListenOffers:
		return cslq.Invokef(d.conn, "c", d.ListenOffers)
	case localListenStatus:
		return cslq.Invokef(d.conn, "c", d.ListenStatus)
	case localUpdatePeer:
		return cslq.Invokef(d.conn, "", d.UpdatePeer)
	}
	return errors.New("protocol violation: unknown command")
}

func nextCommand(d *Dispatcher) (cmd uint8, err error) {
	d.log = warpdrive.NewLogger(d.logPrefix, "(~)")
	err = d.cslq.Decodef("c", &cmd)
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = warpdrive.ErrEnded
		}
		return
	}
	d.log = warpdrive.NewLogger(d.logPrefix, fmt.Sprintf("(%d)", cmd))
	if cmd == cmdClose {
		err = warpdrive.ErrEnded
	}
	return
}

func Invokef2[T1 any, T2 any](r io.Reader, format string, fn func(T1, T2) error) error {
	var v1 T1
	var v2 T2
	if err := cslq.Decode(r, format, &v1, &v2); err != nil {
		return err
	}
	return fn(v1, v2)
}

func Invokef3[T1 any, T2 any, T3 any](r io.Reader, format string, fn func(T1, T2, T3) error) error {
	var v1 T1
	var v2 T2
	var v3 T3
	if err := cslq.Decode(r, format, &v1, &v2, &v3); err != nil {
		return err
	}
	return fn(v1, v2, v3)
}
