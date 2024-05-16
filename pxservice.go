package gopxgrid

import (
	"context"
	"fmt"
)

var (
	ErrInvalidInput = fmt.Errorf("invalid input")
)

type OperationType string

const (
	OperationTypeCreate OperationType = "CREATE"
	OperationTypeUpdate OperationType = "UPDATE"
	OperationTypeDelete OperationType = "DELETE"
)

type SupportedRESTCallDetails struct {
	Name   string   `json:"name"`
	Params []string `json:"params"`
}

type PxGridService interface {
	Name() string
	Nodes() []ServiceNode
	Lookup(ctx context.Context) error
	UpdateSecrets(ctx context.Context) error
	CheckNodes(ctx context.Context) error
	FindProperty(ctx context.Context, property string, nodePick ...ServiceNodePickerFactory) (any, error)
	FindNodeIndexByName(name string) (int, error)
	On(topicProperty string) Subscriber[any]

	GenericRESTCaller
}

type GenericRESTCaller interface {
	AnyREST(call string, payload map[string]any) CallFinalizer[any]
}

var _ PxGridService = (*pxGridService)(nil)

type pxGridService struct {
	name  string
	nodes ServiceNodeSlice
	ctrl  *PxGridConsumer
	log   Logger
}

// Name returns the name of the service
func (s *pxGridService) Name() string {
	return s.name
}

// Lookup retrieves the service nodes from the controller
func (s *pxGridService) Lookup(ctx context.Context) error {
	s.log.Debug("Looking up service", "service", s.name)
	r, err := s.ctrl.ServiceLookup(ctx, s.name)
	if err != nil {
		return err
	}

	s.nodes = r.Services
	return nil
}

// CheckNodes ensures that the service has nodes
func (s *pxGridService) CheckNodes(ctx context.Context) error {
	s.log.Debug("Checking nodes for service", "service", s.name)
	if len(s.nodes) == 0 {
		err := s.Lookup(ctx)
		if err != nil {
			return err
		}
	}

	s.log.Debug("Nodes found for service", "service", s.name, "nodes", len(s.nodes))
	if len(s.nodes) == 0 {
		return ErrServiceUnavailable
	}

	return nil
}

// UpdateNodeSecret retrieves the secret for a node by index
func (s *pxGridService) UpdateNodeSecret(ctx context.Context, idx int) error {
	s.log.Debug("Updating secret for node", "service", s.name, "node", idx)
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
	s.log.Debug("Updating secret for node", "service", s.name, "node", nodeName)
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
func (s *pxGridService) FindProperty(ctx context.Context, property string, nodePick ...ServiceNodePickerFactory) (any, error) {
	err := s.CheckNodes(ctx)
	if err != nil {
		return nil, err
	}

	n := s.orDefaultFactory(nodePick...)(s.nodes)
	for {
		node, more, err := n.PickNode()
		if err != nil {
			return nil, err
		}

		if node.Properties == nil {
			if !more {
				break
			}
			continue
		}

		if v, ok := node.Properties[property]; ok {
			return v, nil
		}

		if !more {
			break
		}
	}

	return nil, fmt.Errorf("property %s not found", property)
}

// Nodes returns copy the service nodes
func (s *pxGridService) Nodes() []ServiceNode {
	nodes := make([]ServiceNode, len(s.nodes))
	copy(nodes, s.nodes)
	return nodes
}

func (s *pxGridService) On(topicProperty string) Subscriber[any] {
	return newSubscriber[any](s, topicProperty, nil)
}

func (s *pxGridService) AnyREST(call string, payload map[string]any) CallFinalizer[any] {
	return newCall[any](s, call, payload, simpleResultMapper[any])
}

func (s *pxGridService) overAll(ctx context.Context, call string, payload any, result any,
	pickNode ...ServiceNodePickerFactory,
) (*Response, error) {
	n := s.orDefaultFactory(pickNode...)(s.nodes)
	for {
		node, more, err := n.PickNode()
		if err != nil {
			return nil, err
		}

		if node.Secret == "" {
			if err := s.UpdateNodeSecretByName(ctx, node.NodeName); err != nil {
				return nil, err
			}
		}

		restBaseURL, ok := node.Properties["restBaseUrl"].(string)
		if !ok {
			if !more {
				break
			}
			continue
		}

		res, err := s.ctrl.RESTRequest(ctx, ensureTrailingSlash(restBaseURL)+call, payload, RESTOptions{
			overridePassword: node.Secret,
			result:           result,
		})
		if err != nil {
			if !more {
				return nil, err
			}
			continue
		}

		return res, nil
	}

	return nil, fmt.Errorf("all nodes failed to %s", call)
}

func (s *pxGridService) call(ctx context.Context, call string, payload any, result any,
	pickNode ...ServiceNodePickerFactory,
) (*Response, error) {
	err := s.CheckNodes(ctx)
	if err != nil {
		return nil, err
	}

	res, err := s.overAll(ctx, call, payload, result, pickNode...)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *pxGridService) orDefaultFactory(f ...ServiceNodePickerFactory) ServiceNodePickerFactory {
	if len(f) > 0 && f[0] != nil {
		return f[0]
	}

	return OrderedNodePicker()
}

func ensureTrailingSlash(s string) string {
	if len(s) == 0 || s[len(s)-1] != '/' {
		return s + "/"
	}

	return s
}
