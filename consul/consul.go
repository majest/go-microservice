package consul

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hashicorp/consul/api"
	"github.com/satori/go.uuid"
)

type Consul struct {
	Client      *api.Client
	serviceName string
	UUID        string
}

var dev string
var port int = 9090

func RegisterService(serviceName string) *Consul {
	c := &Consul{}
	c.Init()
	c.RegisterService(serviceName, uuid.NewV4().String())
	return c
}

func New() *Consul {
	c := &Consul{}
	c.Init()
	return c
}

func (c *Consul) Init() {
	dev = os.Getenv("CONSUL_DEV")
	c.CreateClient()
	c.HandleExit()
}

func (c *Consul) CreateConfig() *api.Config {
	nodeIP, err := c.getNodeIP()

	if err != nil {
		panic(err.Error())
	}

	config := api.DefaultConfig()
	config.Address = nodeIP + ":8500"

	fmt.Println("Config Address:" + config.Address)
	return config
}

func (c *Consul) getNodeIP() (string, error) {

	if dev == "docker-local" || dev == "local" {
		return "192.168.99.101", nil
	}

	awsmeta := ec2metadata.New(session.New(&aws.Config{Region: aws.String("us-east-1")}))
	return awsmeta.GetMetadata("local-ipv4")
}

func (c *Consul) HandleExit() {
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			c.Client.Agent().ServiceDeregister(c.UUID)
			fmt.Println("Deregistering service: " + c.serviceName + ":" + c.UUID)
			os.Exit(0)
		case syscall.SIGTERM:
			c.Client.Agent().ServiceDeregister(c.UUID)
			fmt.Println("Deregistering service: " + c.serviceName + ":" + c.UUID)
			os.Exit(0)
		}
	}()
}

func (c *Consul) CreateClient() {
	client, errc := api.NewClient(c.CreateConfig())
	if errc != nil {
		fmt.Println("Error:" + errc.Error())
		panic(errc.Error())
	}
	c.Client = client
}

func (c *Consul) RegisterService(service, serviceUUID string) {
	c.UUID = serviceUUID
	c.serviceName = service

	fmt.Println("Registering service: " + service + ":" + c.UUID)
	localIP, err := c.getIP()

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Local ip address: %s:%v\n", localIP, port)
	err = c.Client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      c.UUID,
		Name:    service,
		Port:    port,
		Address: localIP,
	})

	if err != nil {
		fmt.Printf("Error while registering service: %s", err.Error())
		os.Exit(1)
	}

}

func (c *Consul) getIP() (string, error) {

	if dev == "local" {
		return "127.0.0.1", nil
	}

	ifaces, err := net.Interfaces()

	if err != nil {
		return "", err
	}

	var ip net.IP

	for _, i := range ifaces {
		addrs, err := i.Addrs()

		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPAddr:
				ip = v.IP
			case *net.IPNet:
				ip = v.IP
			}

			if ip.To4() != nil && ip.String() != "127.0.0.1" {
				return ip.String(), nil
			}
		}
	}

	return "", errors.New("Could not find local ip address")
}
