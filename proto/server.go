package proto

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/go-warpdrive"
	"log"
)

type Server struct {
	warpdrive.Service
	ctx     context.Context
	localId id.Identity
}

func (s *Server) String() string {
	return "[warpdrive]"
}

func (s *Server) Run(ctx context.Context) (err error) {

	s.ctx = ctx

	if s.localId, err = astral.Resolve("localnode"); err != nil {
		return fmt.Errorf("cannot resolve local id: %v", err)
	}

	finish := s.Start(ctx)

	s.Peer().Fetch()

	dispatchers := map[string]func(d *Dispatcher) error{
		warpdrive.Port:    Dispatch,
		warpdrive.PortCli: Cli,
	}

	for port, dispatcher := range dispatchers {
		if err = s.register(port, dispatcher); err != nil {
			return warpdrive.Error(err, "Cannot register", port)
		}
	}

	<-s.ctx.Done()
	<-finish
	return
}

func (s *Server) register(
	query string,
	dispatch func(d *Dispatcher) error,
) (err error) {
	port, err := astral.Register(query)
	if err != nil {
		return
	}

	// Serve handlers
	go func() {
		requestId := uint(0)
		for request := range port.QueryCh() {
			requestId = requestId + 1
			go func(request *astral.QueryData, requestId uint) {
				s.Job().Add(1)
				defer s.Job().Done()

				conn, err := request.Accept()
				defer conn.Close()

				if err != nil {
					err = warpdrive.Error(err, "Cannot accept warpdrive connection")
					log.Println(err)
					return
				}

				logPrefix := fmt.Sprint("[WARPDRIVE] ", query, ":", requestId)

				callerId := request.RemoteIdentity()
				if callerId.IsZero() {
					callerId = s.localId
				}

				authorized := callerId.IsEqual(s.localId)

				_ = NewDispatcher(
					logPrefix,
					callerId.String(),
					authorized,
					s.ctx,
					conn,
					s.Service,
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
