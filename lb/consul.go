package lb

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
)

type Consul struct {
	client      *api.Client
	serviceName string
	UUID        string
}

var dev bool

func (c *Consul) Init(devenv bool) {
	dev = devenv
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
	if dev {
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
			c.client.Agent().ServiceDeregister(c.UUID)
			fmt.Println("Deregistering service: " + c.serviceName + ":" + c.UUID)
			os.Exit(0)
		case syscall.SIGTERM:
			c.client.Agent().ServiceDeregister(c.UUID)
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
	c.client = client
}

func (c *Consul) RegisterService(service, serviceUUID string) {
	c.UUID = serviceUUID
	c.serviceName = service

	fmt.Println("Registering service: " + service + ":" + c.UUID)
	localIP, err := c.getIP()

	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Local ip address: " + localIP)
	err = c.client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      c.UUID,
		Name:    service,
		Port:    8080,
		Address: localIP,
	})

	if err != nil {
		fmt.Printf("Error while registering service: %s", err.Error())
		os.Exit(1)
	}

}

func (c *Consul) getIP() (string, error) {
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
