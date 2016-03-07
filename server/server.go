package server

import (
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/majest/go-microservice/service"
	"github.com/majest/go-microservice/transport"
	"golang.org/x/net/context"
)

func CreateServer(svc service.Service) *httptransport.Server {
	return httptransport.NewServer(
		context.Background(),
		transport.MakeEndpoint(svc),
		transport.DecodeRequest,
		transport.EncodeResponse,
	)
}
