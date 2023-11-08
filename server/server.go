package warpdrived

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/go-warpdrive/adapter"
	"github.com/cryptopunkscc/go-warpdrive/proto"
	"github.com/cryptopunkscc/go-warpdrive/service"
	"log"
)

type Server struct {
	service.Component
	ctx     context.Context
	localId id.Identity
}

func (s *Server) String() string {
	return "[warpdrive]"
}

func (s *Server) Run(ctx context.Context, api adapter.Api) (err error) {

	s.ctx = ctx
	s.Api = api

	if s.localId, err = s.Resolve("localnode"); err != nil {
		return fmt.Errorf("cannot resolve local id: %v", err)
	}

	setupCore(&s.Component)

	finish := service.OfferUpdates(s.Component).Start(s.ctx)

	s.Peer().Fetch()

	dispatchers := map[string]func(d *proto.Dispatcher) error{
		proto.Port:    proto.Dispatch,
		proto.PortCli: proto.Cli,
	}

	for port, dispatcher := range dispatchers {
		if err = s.register(port, dispatcher); err != nil {
			return proto.Error(err, "Cannot register", port)
		}
	}

	<-s.ctx.Done()
	<-finish
	return
}

func (s *Server) register(
	query string,
	dispatch func(d *proto.Dispatcher) error,
) (err error) {
	port, err := s.Api.Register(query)
	if err != nil {
		return
	}

	// Serve handlers
	go func() {
		requestId := uint(0)
		for request := range port.Next() {
			requestId = requestId + 1
			go func(request adapter.Request, requestId uint) {
				s.Job.Add(1)
				defer s.Job.Done()

				conn, err := request.Accept()
				defer conn.Close()

				if err != nil {
					err = proto.Error(err, "Cannot accept warpdrive connection")
					log.Println(err)
					return
				}

				logPrefix := fmt.Sprint("[WARPDRIVE] ", query, ":", requestId)

				callerId := request.Caller()
				if callerId.IsZero() {
					callerId = s.localId
				}

				authorized := callerId.IsEqual(s.localId)

				_ = proto.NewDispatcher(
					logPrefix,
					callerId.String(),
					authorized,
					s.ctx,
					s.Api,
					conn,
					s.Component,
					s.Job,
				).Serve(dispatch)
			}(request, requestId)
		}
	}()
	go func() {
		<-s.ctx.Done()
		port.Close()
	}()
	return
}
