package proto

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/go-warpdrive/adapter"
	"io"
	"log"
	"sync"
)

type Dispatcher struct {
	logPrefix  string
	callerId   string
	authorized bool

	ctx  context.Context
	api  adapter.Api
	conn io.ReadWriteCloser

	srv Service
	job *sync.WaitGroup

	cslq *cslq.Endec
	log  *log.Logger
}

func NewDispatcher(
	logPrefix string,
	callerId string,
	authorized bool,
	ctx context.Context,
	api adapter.Api,
	conn io.ReadWriteCloser,
	srv Service,
	job *sync.WaitGroup,
) *Dispatcher {
	return &Dispatcher{
		logPrefix:  logPrefix,
		callerId:   callerId,
		authorized: authorized,
		ctx:        ctx,
		api:        api,
		conn:       conn,
		srv:        srv,
		job:        job,
		cslq:       cslq.NewEndec(conn),
		log:        NewLogger(logPrefix),
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
	if errors.Is(err, errEnded) {
		d.log.Println("End")
		err = nil
	}
	if err != nil {
		d.log.Println(Error(err, "Failed"))
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
	d.log = NewLogger(d.logPrefix, "(~)")
	err = d.cslq.Decode("c", &cmd)
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = errEnded
		}
		return
	}
	d.log = NewLogger(d.logPrefix, fmt.Sprintf("(%d)", cmd))
	if cmd == cmdClose {
		err = errEnded
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
