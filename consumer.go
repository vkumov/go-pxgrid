package gopxgrid

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type PxGridConsumer struct {
	cfg *PxGridConfig
	svc *transport

	ancConfig        ANCConfig
	endpointAsset    EndpointAsset
	mdm              MDM
	profilerConfig   ProfilerConfiguration
	radiusFailure    RadiusFailure
	sessionDirectory SessionDirectory
	systemHealth     SystemHealth
	trustsecConfig   TrustSecConfiguration
	trustsecSxp      TrustSecSXP
	trustsec         TrustSec

	pubsubs     map[string]PubSub
	pubsubMutex sync.RWMutex
}

var (
	ErrNoHosts            = errors.New("no hosts available")
	ErrServiceUnavailable = errors.New("service unavailable")
)

func mergeWithDefaultConfig(cfg *PxGridConfig) *PxGridConfig {
	if cfg == nil {
		cfg = &PxGridConfig{}
	}

	if cfg.DNS.FamilyStrategy == IPUnknown {
		cfg.DNS.FamilyStrategy = DefaultINETFamilyStrategy
	}

	if cfg.Logger == nil {
		cfg.Logger = FromSlog(slog.Default())
	}

	return cfg
}

func NewPxGridConsumer(cfg *PxGridConfig) (*PxGridConsumer, error) {
	if cfg == nil {
		return nil, errors.New("invalid config")
	}

	c := &PxGridConsumer{
		cfg: mergeWithDefaultConfig(cfg),
		svc: newTransport(cfg),
	}

	c.ancConfig = NewPxGridANCConfig(c)
	c.endpointAsset = NewPxGridEndpointAsset(c)
	c.mdm = NewPxGridMDM(c)
	c.profilerConfig = NewPxGridProfilerConfiguration(c)
	c.radiusFailure = NewPxGridRadiusFailure(c)
	c.sessionDirectory = NewPxGridSessionDirectory(c)
	c.systemHealth = NewPxGridSystemHealth(c)
	c.trustsecConfig = NewPxGridTrustSecConfiguration(c)
	c.trustsecSxp = NewPxGridTrustSecSXP(c)
	c.trustsec = NewPxGridTrustSec(c)

	return c, nil
}

type RESTOptions struct {
	overrideUsername string
	overridePassword string
	noAuth           bool
	result           any
}

func (c *PxGridConsumer) RESTRequest(ctx context.Context, fullURL string, payload any, ops RESTOptions) (*Response, error) {
	req := c.svc.NewRequest(ctx)
	if ops.noAuth {
		req.NoAuth()
	} else {
		if ops.overrideUsername != "" {
			req.SetUsername(ops.overrideUsername)
		}
		if ops.overridePassword != "" {
			req.SetPassword(ops.overridePassword)
		}
	}
	if ops.result != nil {
		req.SetResult(ops.result)
	}
	if c.svc.tls.CA != nil {
		req.SetRootCAs(c.svc.tls.CA)
	} else {
		sys, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		req.SetRootCAs(sys)
	}

	res, err := req.Post(fullURL, payload)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *PxGridConsumer) controlRest(ctx context.Context, urlControl string, payload any, ops RESTOptions) (*Response, error) {
	for _, n := range c.cfg.Hosts {
		port := 8910
		if n.ControlPort != 0 {
			port = n.ControlPort
		}

		fullURL := "https://" + fmt.Sprintf("%s:%d", n.Host, port) + "/pxgrid/control/" + urlControl
		res, err := c.RESTRequest(ctx, fullURL, payload, ops)
		if err != nil {
			continue
		}
		return res, nil
	}

	return nil, ErrNoHosts
}

func (c *PxGridConsumer) ANCConfig() ANCConfig {
	return c.ancConfig
}

func (c *PxGridConsumer) EndpointAsset() EndpointAsset {
	return c.endpointAsset
}

func (c *PxGridConsumer) MDM() MDM {
	return c.mdm
}

func (c *PxGridConsumer) ProfilerConfiguration() ProfilerConfiguration {
	return c.profilerConfig
}

func (c *PxGridConsumer) PubSub(service string) PubSub {
	c.pubsubMutex.Lock()
	defer c.pubsubMutex.Unlock()

	if c.pubsubs == nil {
		c.pubsubs = make(map[string]PubSub)
	}

	svc, ok := c.pubsubs[service]
	if !ok {
		svc = NewPxGridPubSub(c, service)
		c.pubsubs[service] = svc
	}

	return svc
}

func (c *PxGridConsumer) RadiusFailure() RadiusFailure {
	return c.radiusFailure
}

func (c *PxGridConsumer) SessionDirectory() SessionDirectory {
	return c.sessionDirectory
}

func (c *PxGridConsumer) SystemHealth() SystemHealth {
	return c.systemHealth
}

func (c *PxGridConsumer) TrustSecConfiguration() TrustSecConfiguration {
	return c.trustsecConfig
}

func (c *PxGridConsumer) TrustSecSXP() TrustSecSXP {
	return c.trustsecSxp
}

func (c *PxGridConsumer) TrustSec() TrustSec {
	return c.trustsec
}
