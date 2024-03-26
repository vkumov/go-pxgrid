package gopxgrid

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
)

var (
	ErrCreateForbidden      = errors.New("create account forbidden")
	ErrCreateConflict       = errors.New("create account conflict")
	ErrActivateUnauthorized = errors.New("activate account unauthorized")
)

type (
	AccountCreateResponse struct {
		NodeName string `json:"nodeName"`
		Password string `json:"password"`
	}

	AccountActivateResponse struct {
		AccountState string `json:"accountState"`
		Version      string `json:"version"`
	}

	ServiceLookupResponse struct {
		Services []ServiceNode `json:"services"`
	}

	Controller interface {
		RESTRequest(ctx context.Context, fullURL string, payload any, ops RESTOptions) (*Response, error)
		AccountCreate(ctx context.Context) (AccountCreateResponse, error)
		AccountActivate(ctx context.Context) error
		ServiceLookup(ctx context.Context, svc string) (ServiceLookupResponse, error)
		AccessSecret(ctx context.Context, peerNodeName string) (string, error)
	}
)

func (c *PxGridConsumer) Control() Controller {
	return c
}

func (c *PxGridConsumer) AccountCreate(ctx context.Context) (AccountCreateResponse, error) {
	payload := map[string]interface{}{
		"nodeName": c.cfg.NodeName,
	}

	res, err := c.controlRest(ctx, "AccountCreate", payload, RESTOptions{
		noAuth: true,
		result: &AccountCreateResponse{},
	})
	if err != nil {
		return AccountCreateResponse{}, err
	}
	if res.StatusCode == 403 || res.StatusCode == 503 {
		return AccountCreateResponse{}, ErrCreateForbidden
	}
	if res.StatusCode == 409 {
		return AccountCreateResponse{}, ErrCreateConflict
	}
	if res.StatusCode > 299 {
		return AccountCreateResponse{}, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return *(res.Result.(*AccountCreateResponse)), nil
}

func (c *PxGridConsumer) AccountActivate(ctx context.Context) error {
	payload := map[string]interface{}{}
	if c.cfg.Description != "" {
		payload["description"] = c.cfg.Description
	}

	res, err := c.controlRest(ctx, "AccountActivate", payload, RESTOptions{})
	if err != nil {
		return err
	}

	if res.StatusCode == 401 {
		return ErrActivateUnauthorized
	}
	if res.StatusCode > 299 {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return nil
}

func (c *PxGridConsumer) ServiceLookup(ctx context.Context, svc string) (ServiceLookupResponse, error) {
	payload := map[string]interface{}{
		"name": svc,
	}

	res, err := c.controlRest(ctx, "ServiceLookup", payload, RESTOptions{
		result: &ServiceLookupResponse{},
	})
	if err != nil {
		return ServiceLookupResponse{}, err
	}

	return *(res.Result.(*ServiceLookupResponse)), nil
}

func (c *PxGridConsumer) AccessSecret(ctx context.Context, peerNodeName string) (string, error) {
	payload := map[string]interface{}{
		"peerNodeName": peerNodeName,
	}

	type AccessSecretResponse struct {
		Secret string `json:"secret"`
	}

	res, err := c.controlRest(ctx, "AccessSecret", payload, RESTOptions{
		result: &AccessSecretResponse{},
	})
	if err != nil {
		return "", err
	}

	got := res.Result.(*AccessSecretResponse)
	return got.Secret, nil
}

func (c *PxGridConsumer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return c.svc.DialContext(ctx, network, addr)
}

func (c *PxGridConsumer) ClientTLSConfig() *tls.Config {
	return c.svc.ClientTLSConfig()
}
