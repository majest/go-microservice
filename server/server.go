package server

import (
	"io"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/loadbalancer/consul"
	"github.com/go-kit/kit/log"
	kitratelimit "github.com/go-kit/kit/ratelimit"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	jujuratelimit "github.com/juju/ratelimit"
	clb "github.com/majest/go-microservice/consul"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
)

type ClientConfig struct {
	ServiceName     string
	GRPCSettings    []grpc.DialOption
	MaxQPS          int
	MaxAttempts     int
	MaxTime         time.Duration
	BreakerSettings gobreaker.Settings
	ConsulIP        string
	ConsulPort      int
}

// MakeEndpoint wires up a gRPC endpoint with Go kit tooling
func MakeEndpoint(method string, logger log.Logger, enc kitgrpc.EncodeRequestFunc, dec kitgrpc.DecodeResponseFunc, reply interface{}, config *ClientConfig) endpoint.Endpoint {
	// generate a load balancer factory for endpoints
	factoryFunc := func(instance string) (endpoint.Endpoint, io.Closer, error) {
		// Try to establish connection to gRPC Service
		cc, err := grpc.Dial(instance, config.GRPCSettings...)
		if err != nil {
			return nil, nil, err
		}
		// Create Go Kit gRPC Client Endpoint
		var e endpoint.Endpoint
		e = kitgrpc.NewClient(
			cc,
			config.ServiceName,
			method,
			enc,
			dec,
			reply,
		//    kitgrpc.SetClientBefore(traceFunc),
		).Endpoint()

		// Wrap endpoint with a Circuit Breaker
		e = circuitbreaker.Gobreaker(
			gobreaker.NewCircuitBreaker(config.BreakerSettings),
		)(e)
		// Wrap endpoint with a Rate Limiter
		e = kitratelimit.NewTokenBucketLimiter(
			jujuratelimit.NewBucketWithRate(
				float64(config.MaxQPS),
				int64(config.MaxQPS),
			),
		)(e)
		// return Endpoint to Endpoint Cache
		return e, nil, nil
	}
	// create the needed publisherFactory (Zookeeper based)
	logger = log.NewContext(logger).With("component", "loadbalancer/consul")
	//    zkClient, _ := zk.NewClient(zkhosts, logger)
	cclient := consul.NewClient(clb.New(&clb.Config{NodeIp: config.ConsulIP, NodePort: config.ConsulPort}).Client)
	//p, _ := zk.NewPublisher(zkClient, path, factoryFunc, logger)
	//p, _ := consul.NewPublisher(cclient, factoryFunc, logger, config.ServiceName)
	p, _ := consul.NewPublisher(cclient, factoryFunc, logger, config.ServiceName)
	// create a Round Robin loadbalancer for the discovered endpoints
	lb := loadbalancer.NewRoundRobin(p)
	// wrap our loadbalancer with retry and timeout logic
	e := loadbalancer.Retry(config.MaxAttempts, config.MaxTime, lb)
	// // annotate our endpoint with zipkin tracing
	// e = zipkin.AnnotateClient(spanFunc, config.Collector)(e)
	// return the fully wrapped & decorated endpoint
	return e
}
