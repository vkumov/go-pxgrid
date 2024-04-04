package gopxgrid

import (
	"crypto/tls"
	"crypto/x509"
)

type INETFamilyStrategy int

const (
	IPUnknown INETFamilyStrategy = iota
	IPv4
	IPv46
	IPv64
	IPv6
)

var DefaultINETFamilyStrategy = IPv46

type DNSConfig struct {
	Server         string
	FamilyStrategy INETFamilyStrategy
}

type Host struct {
	Host        string
	ControlPort int
}

type AuthConfig struct {
	Username string
	Password string
}

type TLSConfig struct {
	ClientCertificate *tls.Certificate
	InsecureTLS       bool
	CA                *x509.CertPool
}

type PxGridConfig struct {
	Hosts       []Host
	Auth        AuthConfig
	NodeName    string
	Description string
	TLS         TLSConfig
	DNS         DNSConfig
	Logger      Logger
}

func NewPxGridConfig() *PxGridConfig {
	return &PxGridConfig{
		DNS: DNSConfig{
			FamilyStrategy: DefaultINETFamilyStrategy,
		},
	}
}

func (c *PxGridConfig) AddHost(host string, controlPort int) *PxGridConfig {
	c.Hosts = append(c.Hosts, Host{
		Host:        host,
		ControlPort: controlPort,
	})
	return c
}

func (c *PxGridConfig) SetAuth(username, password string) *PxGridConfig {
	c.Auth = AuthConfig{
		Username: username,
		Password: password,
	}
	return c
}

func (c *PxGridConfig) SetNodeName(name string) *PxGridConfig {
	c.NodeName = name
	return c
}

func (c *PxGridConfig) SetDescription(desc string) *PxGridConfig {
	c.Description = desc
	return c
}

func (c *PxGridConfig) SetLogger(logger Logger) *PxGridConfig {
	c.Logger = logger
	return c
}

func (c *PxGridConfig) SetClientCertificate(cert *tls.Certificate) *PxGridConfig {
	c.TLS.ClientCertificate = cert
	return c
}

func (c *PxGridConfig) SetInsecureTLS(insecure bool) *PxGridConfig {
	c.TLS.InsecureTLS = insecure
	return c
}

func (c *PxGridConfig) SetCA(ca *x509.CertPool) *PxGridConfig {
	c.TLS.CA = ca
	return c
}

func (c *PxGridConfig) SetDNS(server string, family INETFamilyStrategy) *PxGridConfig {
	c.DNS.Server = server
	c.DNS.FamilyStrategy = family
	return c
}
