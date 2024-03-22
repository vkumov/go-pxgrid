package gopxgrid

import (
	"context"
	"fmt"
)

type PxGridConsumer struct {
	cfg *PxGridConfig
	svc *service
}

var (
	ErrNoHosts = fmt.Errorf("no hosts available")
)

func NewPxGridConsumer(cfg *PxGridConfig) *PxGridConsumer {
	return &PxGridConsumer{
		cfg: cfg,
		svc: newService(cfg),
	}
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
