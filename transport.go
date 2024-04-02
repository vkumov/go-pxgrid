package gopxgrid

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/go-resty/resty/v2"
)

type transport struct {
	client   *resty.Client
	tls      *TLSConfig
	dns      *DNSConfig
	resolver *net.Resolver
	auth     AuthConfig

	tlsMutex sync.RWMutex
}

func dnsCfg(cfg *DNSConfig) *DNSConfig {
	if cfg == nil {
		return &DNSConfig{
			FamilyStrategy: DefaultINETFamilyStrategy,
		}
	}
	if cfg.FamilyStrategy == IPUnknown {
		cfg.FamilyStrategy = DefaultINETFamilyStrategy
	}
	return cfg
}

func tlsCfg(cfg *TLSConfig) *TLSConfig {
	if cfg == nil {
		return &TLSConfig{
			InsecureTLS: true,
		}
	}

	return cfg
}

func newTransport(cfg *PxGridConfig) *transport {
	s := &transport{
		client: resty.New(),
		tls:    tlsCfg(&cfg.TLS),
		dns:    dnsCfg(&cfg.DNS),
		auth:   cfg.Auth,
	}

	if s.dns.Server != "" {
		host := s.dns.Server
		if strings.Index(host, ":") == -1 {
			host = host + ":53"
		}

		s.resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "udp", host)
			},
		}
	}

	s.client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: s.tls.InsecureTLS})

	s.client.SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	})

	return s
}

func (s *transport) getOneIPAddr(addrs []net.IPAddr, err error) (net.IPAddr, error) {
	if err != nil {
		return net.IPAddr{}, err
	}

	switch s.dns.FamilyStrategy {
	case IPv4:
		for _, i := range addrs {
			if i.IP.To4() != nil {
				return i, nil
			}
		}
		return net.IPAddr{}, &net.AddrError{Err: "no IPv4 address found", Addr: ""}
	case IPv6:
		for _, i := range addrs {
			if i.IP.To4() == nil {
				return i, nil
			}
		}
		return net.IPAddr{}, &net.AddrError{Err: "no IPv6 address found", Addr: ""}
	case IPv46:
		var firstIPv6 net.IPAddr
		for _, i := range addrs {
			if i.IP.To4() != nil {
				return i, nil
			}
			if firstIPv6.IP == nil {
				firstIPv6 = i
			}
		}
		if firstIPv6.IP != nil {
			return firstIPv6, nil
		}
		return net.IPAddr{}, &net.AddrError{Err: "no IPv4 or IPv6 address found", Addr: ""}
	case IPv64:
		var firstIPv4 net.IPAddr
		for _, i := range addrs {
			if i.IP.To4() == nil {
				return i, nil
			}
			if firstIPv4.IP == nil {
				firstIPv4 = i
			}
		}
		if firstIPv4.IP != nil {
			return firstIPv4, nil
		}
		return net.IPAddr{}, &net.AddrError{Err: "no IPv4 or IPv6 address found", Addr: ""}
	}

	return net.IPAddr{}, &net.AddrError{Err: "unknown strategy", Addr: ""}
}

func (s *transport) ResolveHost(ctx context.Context, host string) (net.IPAddr, error) {
	if s.resolver == nil {
		return s.getOneIPAddr(net.DefaultResolver.LookupIPAddr(ctx, host))
	}

	return s.getOneIPAddr(s.resolver.LookupIPAddr(ctx, host))
}

func (s *transport) UpdateClientCertificate(cert *tls.Certificate) {
	s.tlsMutex.Lock()
	defer s.tlsMutex.Unlock()

	s.tls.ClientCertificate = cert
}

func (s *transport) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{
		Resolver: s.resolver,
	}
	return dialer.DialContext(ctx, network, addr)
}

func (s *transport) ClientTLSConfig() *tls.Config {
	s.tlsMutex.RLock()
	defer s.tlsMutex.RUnlock()

	tls := &tls.Config{InsecureSkipVerify: s.tls.InsecureTLS}
	if s.tls.ClientCertificate != nil {
		tls.Certificates = append(tls.Certificates, *s.tls.ClientCertificate)
	}
	if s.tls.CA != nil {
		tls.RootCAs = s.tls.CA.Clone()
	}

	return tls
}

func (s *transport) NewRequest(ctx context.Context) *Request {
	clonedAuth := s.auth
	return &Request{
		s:       s,
		ctx:     ctx,
		auth:    &clonedAuth,
		rootCAs: nil,
		tls:     s.tls,
		client:  s.client.Clone(),
	}
}

type (
	Request struct {
		s       *transport
		ctx     context.Context
		auth    *AuthConfig
		rootCAs *x509.CertPool
		tls     *TLSConfig
		client  *resty.Client
		result  interface{}
	}

	Response struct {
		StatusCode int
		Body       []byte
		Result     interface{}
	}
)

func (r *Request) getTLSClientConfig(serverName string) *tls.Config {
	cfg := &tls.Config{
		InsecureSkipVerify: r.tls.InsecureTLS,
	}
	if r.tls.ClientCertificate != nil {
		cfg.Certificates = []tls.Certificate{*r.tls.ClientCertificate}
	}

	cfg.ServerName = serverName
	if r.rootCAs != nil {
		cfg.RootCAs = r.rootCAs
	}
	return cfg
}

func (r *Request) getAuth() (string, string) {
	if r.auth == nil {
		return "", ""
	}

	return r.auth.Username, r.auth.Password
}

// NoAuth disables authentication for the request.
func (r *Request) NoAuth() *Request {
	r.auth = nil
	return r
}

// SetPassword sets the password for the request.
func (r *Request) SetPassword(password string) *Request {
	if r.auth == nil {
		r.auth = &AuthConfig{}
	}
	r.auth.Password = password
	return r
}

// SetUsername sets the username for the request.
func (r *Request) SetUsername(username string) *Request {
	if r.auth == nil {
		r.auth = &AuthConfig{}
	}
	r.auth.Username = username
	return r
}

// SetRootCAs sets the root CAs for the request.
func (r *Request) SetRootCAs(rootCAs *x509.CertPool) *Request {
	r.rootCAs = rootCAs
	return r
}

// SetTLSConfig sets the TLS configuration for the request.
func (r *Request) SetTLSConfig(tls *TLSConfig) *Request {
	r.tls = tls
	return r
}

func (r *Request) SetResult(result interface{}) *Request {
	r.result = result
	return r
}

// Post sends a POST request to the specified URL with the given payload.
func (r *Request) Post(u string, payload interface{}) (*Response, error) {
	o, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	hostname := o.Hostname()
	ip, err := r.s.ResolveHost(r.ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host: %w", err)
	}

	port := 8910
	if o.Port() != "" {
		port, err = strconv.Atoi(o.Port())
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
	}

	r.client.SetTLSClientConfig(r.getTLSClientConfig(hostname))
	req := r.client.R()

	if r.auth != nil {
		req.SetBasicAuth(r.getAuth())
	}
	if r.result != nil {
		req.SetResult(r.result)
	}

	resp, err := req.
		SetHeader("Host", hostname).
		SetBody(payload).
		Post(fmt.Sprintf("https://%s:%d%s", ip.String(), port, o.Path))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	done := Response{
		Body:       resp.Body(),
		StatusCode: resp.StatusCode(),
	}
	if r.result != nil {
		done.Result = resp.Result()
	}

	return &done, nil
}
