package gopxgrid

import (
	"context"
	"fmt"
)

var (
	ErrInvalidInput = fmt.Errorf("invalid input")
)

type SupportedRESTCallDetails struct {
	Name   string   `json:"name"`
	Params []string `json:"params"`
}

type PxGridService interface {
	Name() string
	Lookup(ctx context.Context) error
	UpdateSecrets(ctx context.Context) error
	CheckNodes(ctx context.Context) error
	FindProperty(ctx context.Context, property string, onlyNodes ...int) (any, error)
	FindNodeIndexByName(name string) (int, error)
}

var _ PxGridService = (*pxGridService)(nil)

type pxGridService struct {
	name  string
	nodes []ServiceNode
	ctrl  Controller
}

// Name returns the name of the service
func (s *pxGridService) Name() string {
	return s.name
}

// Lookup retrieves the service nodes from the controller
func (s *pxGridService) Lookup(ctx context.Context) error {
	r, err := s.ctrl.ServiceLookup(ctx, s.name)
	if err != nil {
		return err
	}

	s.nodes = r.Services
	return nil
}

// CheckNodes ensures that the service has nodes
func (s *pxGridService) CheckNodes(ctx context.Context) error {
	if len(s.nodes) == 0 {
		err := s.Lookup(ctx)
		if err != nil {
			return err
		}
	}

	if len(s.nodes) == 0 {
		return ErrNoHosts
	}

	return nil
}

// UpdateNodeSecret retrieves the secret for a node by index
func (s *pxGridService) UpdateNodeSecret(ctx context.Context, idx int) error {
	if idx < 0 || idx >= len(s.nodes) {
		return fmt.Errorf("invalid node index %d", idx)
	}

	secret, err := s.ctrl.AccessSecret(ctx, s.nodes[idx].NodeName)
	if err != nil {
		return fmt.Errorf("failed to get secret for node %s: %w", s.nodes[idx].NodeName, err)
	}

	s.nodes[idx].Secret = secret
	return nil
}

// UpdateNodeSecretByName retrieves the secret for a node by name
func (s *pxGridService) UpdateNodeSecretByName(ctx context.Context, nodeName string) error {
	idx, err := s.FindNodeIndexByName(nodeName)
	if err != nil {
		return err
	}

	return s.UpdateNodeSecret(ctx, idx)
}

// UpdateSecrets retrieves the secrets for all nodes
func (s *pxGridService) UpdateSecrets(ctx context.Context) error {
	err := s.CheckNodes(ctx)
	if err != nil {
		return err
	}

	for i := range s.nodes {
		err := s.UpdateNodeSecret(ctx, i)
		if err != nil {
			return err
		}
	}

	return nil
}

// FindNodeIndexByName returns the index of a node by name
func (s *pxGridService) FindNodeIndexByName(name string) (int, error) {
	for i, n := range s.nodes {
		if n.NodeName == name {
			return i, nil
		}
	}

	return -1, fmt.Errorf("node %s not found", name)
}

func (s *pxGridService) getIterateNodes(onlyNodes ...int) []ServiceNode {
	if len(onlyNodes) == 0 {
		return s.nodes
	}

	iterateOver := make([]ServiceNode, 0, len(onlyNodes))
	for _, i := range onlyNodes {
		if i >= 0 && i < len(s.nodes) {
			iterateOver = append(iterateOver, s.nodes[i])
		}
	}

	return iterateOver
}

// FindProperty returns the value of a property
func (s *pxGridService) FindProperty(ctx context.Context, property string, onlyNodes ...int) (any, error) {
	err := s.CheckNodes(ctx)
	if err != nil {
		return nil, err
	}

	for _, n := range s.getIterateNodes(onlyNodes...) {
		if n.Properties == nil {
			continue
		}

		if v, ok := n.Properties[property]; ok {
			return v, nil
		}
	}

	return nil, fmt.Errorf("property %s not found", property)
}

func (s *pxGridService) overAll(ctx context.Context, call string, payload any, result any, onlyNodes ...int) (*Response, error) {
	for _, n := range s.getIterateNodes(onlyNodes...) {
		if n.Secret == "" {
			if err := s.UpdateNodeSecretByName(ctx, n.NodeName); err != nil {
				return nil, err
			}
		}

		restBaseURL, ok := n.Properties["restBaseUrl"].(string)
		if !ok {
			continue
		}

		res, err := s.ctrl.RESTRequest(ctx, ensureTrailingSlash(restBaseURL)+call, payload, RESTOptions{
			overridePassword: n.Secret,
			result:           result,
		})
		if err != nil {
			continue
		}

		return res, nil
	}

	return nil, fmt.Errorf("all nodes failed to %s", call)
}

func (s *pxGridService) call(ctx context.Context, call string, payload any, result any, onlyNodes ...int) (*Response, error) {
	err := s.CheckNodes(ctx)
	if err != nil {
		return nil, err
	}

	res, err := s.overAll(ctx, call, payload, result, onlyNodes...)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func ensureTrailingSlash(s string) string {
	if len(s) == 0 || s[len(s)-1] != '/' {
		return s + "/"
	}

	return s
}

type CallFinalizer[T any] interface {
	Do(ctx context.Context) (T, error)
	DoOnNode(ctx context.Context, node int) (T, error)
	DoOnNodeByName(ctx context.Context, nodeName string) (T, error)
	DoOnNodes(ctx context.Context, nodes ...int) (T, error)
}

type NoResultCallFinalizer interface {
	Do(ctx context.Context) error
	DoOnNode(ctx context.Context, node int) error
	DoOnNodeByName(ctx context.Context, nodeName string) error
	DoOnNodes(ctx context.Context, nodes ...int) error
}

type call[R any] struct {
	svc     *pxGridService
	call    string
	payload any
	result  R
	mapper  func(*Response) (R, error)

	fatal error
}

type noResultCall[T any] struct {
	svc     *pxGridService
	call    string
	payload any
	mapper  func(*Response) error

	fatal error
}

func (c *call[R]) Do(ctx context.Context) (R, error) {
	if c.fatal != nil {
		return c.result, c.fatal
	}

	res, err := c.svc.call(ctx, c.call, c.payload, c.result)
	if err != nil {
		return c.result, err
	}

	if c.mapper != nil {
		return c.mapper(res)
	}

	return res.Result.(R), nil
}

func (c *call[R]) DoOnNode(ctx context.Context, node int) (R, error) {
	if c.fatal != nil {
		return c.result, c.fatal
	}

	res, err := c.svc.call(ctx, c.call, c.payload, c.result, node)
	if err != nil {
		return c.result, err
	}

	if c.mapper != nil {
		return c.mapper(res)
	}

	return res.Result.(R), nil
}

func (c *call[R]) DoOnNodeByName(ctx context.Context, nodeName string) (R, error) {
	if c.fatal != nil {
		return c.result, c.fatal
	}

	idx, err := c.svc.FindNodeIndexByName(nodeName)
	if err != nil {
		return c.result, err
	}

	return c.DoOnNode(ctx, idx)
}

func (c *call[R]) DoOnNodes(ctx context.Context, nodes ...int) (R, error) {
	if c.fatal != nil {
		return c.result, c.fatal
	}

	res, err := c.svc.call(ctx, c.call, c.payload, c.result, nodes...)
	if err != nil {
		return c.result, err
	}

	if c.mapper != nil {
		return c.mapper(res)
	}

	return res.Result.(R), nil
}

func newCall[R any](svc *pxGridService, apiCall string, payload any, mapper func(*Response) (R, error)) CallFinalizer[R] {
	var result R
	return &call[R]{
		svc:     svc,
		call:    apiCall,
		payload: payload,
		result:  result,
		mapper:  mapper,
	}
}

func newFailedCall[R any](err error) CallFinalizer[R] {
	var result R
	return &call[R]{
		result: result,
		fatal:  err,
	}
}

func (c *noResultCall[T]) Do(ctx context.Context) error {
	if c.fatal != nil {
		return c.fatal
	}

	var result T
	res, err := c.svc.call(ctx, c.call, c.payload, result)
	if err != nil {
		return err
	}

	if c.mapper != nil {
		return c.mapper(res)
	}

	return nil
}

func (c *noResultCall[T]) DoOnNode(ctx context.Context, node int) error {
	if c.fatal != nil {
		return c.fatal
	}

	var result T
	res, err := c.svc.call(ctx, c.call, c.payload, result, node)
	if err != nil {
		return err
	}
	if c.mapper != nil {
		return c.mapper(res)
	}
	return nil
}

func (c *noResultCall[T]) DoOnNodeByName(ctx context.Context, nodeName string) error {
	if c.fatal != nil {
		return c.fatal
	}

	idx, err := c.svc.FindNodeIndexByName(nodeName)
	if err != nil {
		return err
	}

	return c.DoOnNode(ctx, idx)
}

func (c *noResultCall[T]) DoOnNodes(ctx context.Context, nodes ...int) error {
	if c.fatal != nil {
		return c.fatal
	}

	var result T
	res, err := c.svc.call(ctx, c.call, c.payload, result, nodes...)
	if err != nil {
		return err
	}

	if c.mapper != nil {
		return c.mapper(res)
	}

	return nil
}

func newNoResultCall[T any](svc *pxGridService, apiCall string, payload any, mapper func(*Response) error) NoResultCallFinalizer {
	return &noResultCall[T]{
		svc:     svc,
		call:    apiCall,
		payload: payload,
		mapper:  mapper,
	}
}

func newFailedNoResultCall(err error) NoResultCallFinalizer {
	return &noResultCall[any]{
		fatal: err,
	}
}

func simpleResultMapper[T any](r *Response) (T, error) {
	if r.StatusCode > 299 {
		var t T
		return t, fmt.Errorf("unexpected status code: %d", r.StatusCode)
	}
	return r.Result.(T), nil
}

func simpleNoResultMapper(r *Response) error {
	if r.StatusCode > 299 {
		return fmt.Errorf("unexpected status code: %d", r.StatusCode)
	}
	return nil
}
