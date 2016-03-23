package consul

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/consul/api"
	"github.com/satori/go.uuid"
)

type Consul struct {
	Client      *api.Client
	Config      *Config
	serviceName string
	UUID        string
}

func RegisterService(serviceName string, config *Config) *Consul {

	if config == nil {
		config = &Config{}
	}

	c := &Consul{}
	config.SetDefaults()
	c.Config = config

	c.init()
	c.RegisterService(serviceName, uuid.NewV4().String())
	return c
}

func New(config *Config) *Consul {
	if config == nil {
		config = &Config{}
	}

	c := &Consul{}
	config.SetDefaults()
	c.Config = config

	c.init()
	return c
}

func (c *Consul) init() {
	c.createClient()
	c.handleExit()
}

func (c *Consul) createApiConfig() *api.Config {

	apiConfig := api.DefaultConfig()
	apiConfig.Address = fmt.Sprintf("%s:%v", c.Config.NodeIp, c.Config.NodePort)
	fmt.Println("Consul Address:" + apiConfig.Address)
	return apiConfig
}

func (c *Consul) handleExit() {
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

func (c *Consul) createClient() {
	client, errc := api.NewClient(c.createApiConfig())
	if errc != nil {
		fmt.Println("Error:" + errc.Error())
		os.Exit(1)
	}
	c.Client = client
}

func (c *Consul) RegisterService(service, serviceUUID string) {
	c.UUID = serviceUUID
	c.serviceName = service

	fmt.Printf("Service address: %s:%v\n", c.Config.ServiceIp, c.Config.ServicePort)
	err := c.Client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      c.UUID,
		Name:    service,
		Port:    c.Config.ServicePort,
		Address: c.Config.ServiceIp,
	})

	if err != nil {
		fmt.Printf("Error while registering service: %s", err.Error())
		os.Exit(1)
	}
}
