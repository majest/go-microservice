package consul

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Config struct {
	ServiceIp   string
	ServicePort int
	NodeIp      string
	NodePort    int
}

const DEFAULT_SERVICE_PORT = 9090
const DEFAULT_CONSUL_PORT = 8500

func (c *Config) SetDefaults() {

	if c.ServicePort == 0 {
		fmt.Printf("Service port: %v\n", c.ServicePort)
		c.ServicePort = DEFAULT_SERVICE_PORT
	}

	// service up, if no provided try to obtain local ip
	if c.ServiceIp == "" {
		serviceIp, err := c.getIP()

		if err != nil {
			fmt.Printf("Could not get local ip: %s\n", err.Error())
			os.Exit(1)
		}

		c.ServiceIp = serviceIp
	}

	if c.NodeIp == "" {
		nodeIp, err := c.getNodeIP()

		// fallback to default ip
		if err != nil {
			fmt.Printf("Could not get consul node ip: %s\n", err.Error())
			os.Exit(1)
		}
		c.NodeIp = nodeIp
	}

	if c.NodePort == 0 {
		c.NodePort = DEFAULT_CONSUL_PORT
	}
}

// try to get ip of the node using ec2 metadata
func (c *Config) getNodeIP() (string, error) {
	awsmeta := ec2metadata.New(session.New(&aws.Config{Region: aws.String("us-east-1")}))
	return awsmeta.GetMetadata("local-ipv4")
}

func (c *Config) getIP() (string, error) {

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
