package server

import (
	"net"

	"github.com/majest/go-microservice/lb"
	"github.com/satori/go.uuid"

	"google.golang.org/grpc"
)

type Config struct {
	Name        string
	Description string
}

type Server struct {
	c         *Config
	ln        net.Listener
	transport *grpc.Server
}

func Init(c *Config) *Server {
	s := &Server{
		c: c,
	}
	s.Init()
	return s
}

func (s *Server) RegisterWithLb() {
	loadbalancer := lb.Consul{}
	loadbalancer.Init()
	loadbalancer.RegisterService(s.c.Name, uuid.NewV4().String())
}

func (s *Server) Init() {
	ln, err := net.Listen("tcp", ":9090")

	if err != nil {
		panic(err.Error())
	}

	s.ln = ln

	s.transport = grpc.NewServer() // uses its own, internal context

}

func (s *Server) Transport() *grpc.Server {
	return s.transport
}

func (s *Server) Run() {
	s.transport.Serve(s.ln)
}
