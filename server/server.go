package server

import (
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/majest/go-microservice/service"
	"github.com/majest/go-microservice/transport"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

var uid string

type Server struct {
	ctx context.Context
}

func (s *Server) init() {
	s.ctx = context.Background()
	uid = uuid.NewV4().String()
}

func (s *Server) CreateServer(svc service.Service) *httptransport.Server {
	return httptransport.NewServer(
		s.ctx,
		transport.MakeEndpoint(svc),
		transport.DecodeRequest,
		transport.EncodeResponse,
	)
}
