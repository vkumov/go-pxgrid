package gopxgrid

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/go-stomp/stomp/v3"
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
	CA                []x509.Certificate

	pool *x509.CertPool
}

type PxGridConfig struct {
	Hosts       []Host
	Auth        AuthConfig
	NodeName    string
	Description string
	TLS         TLSConfig
	DNS         DNSConfig
	Logger      stomp.Logger
}
