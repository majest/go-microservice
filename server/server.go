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
}

func (s *Server) init() {
	uid = uuid.NewV4().String()
}

func (s *Server) CreateServer(svc service.Service) *httptransport.Server {
	return httptransport.NewServer(
		context.Background(),
		transport.MakeEndpoint(svc),
		transport.DecodeRequest,
		transport.EncodeResponse,
	)
}
