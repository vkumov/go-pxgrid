package gopxgrid

import (
	"context"
	"fmt"
)

type (
	FullResponse[R any] struct {
		StatusCode int
		Result     R
		Body       []byte
	}

	NoResultResponse struct {
		StatusCode int
	}

	CallFinalizer[T any] interface {
		Do(ctx context.Context) (FullResponse[T], error)
		DoOnNode(ctx context.Context, node int) (FullResponse[T], error)
		DoOnNodeByName(ctx context.Context, nodeName string) (FullResponse[T], error)
		DoOnNodes(ctx context.Context, nodes ...int) (FullResponse[T], error)
	}

	NoResultCallFinalizer interface {
		Do(ctx context.Context) (NoResultResponse, error)
		DoOnNode(ctx context.Context, node int) (NoResultResponse, error)
		DoOnNodeByName(ctx context.Context, nodeName string) (NoResultResponse, error)
		DoOnNodes(ctx context.Context, nodes ...int) (NoResultResponse, error)
	}

	call[R any] struct {
		svc     *pxGridService
		call    string
		payload any
		result  R
		mapper  func(*Response) (R, error)

		fatal error
	}

	noResultCall[T any] struct {
		svc     *pxGridService
		call    string
		payload any
		mapper  func(*Response) error

		fatal error
	}
)

func (c *call[R]) Do(ctx context.Context) (FullResponse[R], error) {
	if c.fatal != nil {
		return c.returnError(c.fatal)
	}

	res, err := c.svc.call(ctx, c.call, c.payload, c.result)
	if err != nil {
		return c.returnError(err)
	}

	return c.returnResult(res)
}

func (c *call[R]) DoOnNode(ctx context.Context, node int) (FullResponse[R], error) {
	if c.fatal != nil {
		return c.returnError(c.fatal)
	}

	res, err := c.svc.call(ctx, c.call, c.payload, c.result, IndexNodePicker(node))
	if err != nil {
		return c.returnError(err)
	}

	return c.returnResult(res)
}

func (c *call[R]) DoOnNodeByName(ctx context.Context, nodeName string) (FullResponse[R], error) {
	if c.fatal != nil {
		return c.returnError(c.fatal)
	}

	idx, err := c.svc.FindNodeIndexByName(nodeName)
	if err != nil {
		return c.returnError(err)
	}

	return c.DoOnNode(ctx, idx)
}

func (c *call[R]) DoOnNodes(ctx context.Context, nodes ...int) (FullResponse[R], error) {
	if c.fatal != nil {
		return c.returnError(c.fatal)
	}

	res, err := c.svc.call(ctx, c.call, c.payload, c.result, IndexNodePicker(nodes...))
	if err != nil {
		return c.returnError(err)
	}

	return c.returnResult(res)
}

func (c *call[R]) returnError(err error) (FullResponse[R], error) {
	return FullResponse[R]{Result: c.result}, err
}

func (c *call[R]) returnResult(res *Response) (FullResponse[R], error) {
	if c.mapper != nil {
		mapped, err := c.mapper(res)
		return FullResponse[R]{
			Result:     mapped,
			StatusCode: res.StatusCode,
			Body:       res.Body,
		}, err
	}

	if res.Result == nil {
		return FullResponse[R]{
			Result:     c.result,
			StatusCode: res.StatusCode,
			Body:       res.Body,
		}, nil
	}

	return FullResponse[R]{
		Result:     res.Result.(R),
		StatusCode: res.StatusCode,
		Body:       res.Body,
	}, nil
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

func (c *noResultCall[T]) Do(ctx context.Context) (NoResultResponse, error) {
	if c.fatal != nil {
		return c.returnError(c.fatal)
	}

	var result T
	res, err := c.svc.call(ctx, c.call, c.payload, result)
	if err != nil {
		return c.returnError(err)
	}

	return c.returnResult(res)
}

func (c *noResultCall[T]) DoOnNode(ctx context.Context, node int) (NoResultResponse, error) {
	if c.fatal != nil {
		return c.returnError(c.fatal)
	}

	var result T
	res, err := c.svc.call(ctx, c.call, c.payload, result, IndexNodePicker(node))
	if err != nil {
		return c.returnError(err)
	}

	return c.returnResult(res)
}

func (c *noResultCall[T]) DoOnNodeByName(ctx context.Context, nodeName string) (NoResultResponse, error) {
	if c.fatal != nil {
		return c.returnError(c.fatal)
	}

	idx, err := c.svc.FindNodeIndexByName(nodeName)
	if err != nil {
		return c.returnError(err)
	}

	return c.DoOnNode(ctx, idx)
}

func (c *noResultCall[T]) DoOnNodes(ctx context.Context, nodes ...int) (NoResultResponse, error) {
	if c.fatal != nil {
		return c.returnError(c.fatal)
	}

	var result T
	res, err := c.svc.call(ctx, c.call, c.payload, result, IndexNodePicker(nodes...))
	if err != nil {
		return c.returnError(err)
	}

	return c.returnResult(res)
}

func (c *noResultCall[T]) returnError(err error) (NoResultResponse, error) {
	return NoResultResponse{}, err
}

func (c *noResultCall[T]) returnResult(res *Response) (NoResultResponse, error) {
	if c.mapper != nil {
		err := c.mapper(res)
		return NoResultResponse{StatusCode: res.StatusCode}, err
	}

	return NoResultResponse{StatusCode: res.StatusCode}, nil
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
	if r.StatusCode == 204 {
		var t T
		return t, nil
	}
	if r.Result == nil {
		var t T
		return t, nil
	}
	return r.Result.(T), nil
}

func simpleNoResultMapper(r *Response) error {
	if r.StatusCode > 299 {
		return fmt.Errorf("unexpected status code: %d", r.StatusCode)
	}
	return nil
}
