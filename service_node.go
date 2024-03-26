package gopxgrid

import (
	"errors"
	"math/rand"
)

type (
	ServiceNode struct {
		Name       string                 `json:"name"`
		NodeName   string                 `json:"nodeName"`
		Properties map[string]interface{} `json:"properties"`
		Secret     string                 `json:"-"`
	}

	ServiceNodeSlice []ServiceNode
)

var (
	ErrPropertyNotFound  = errors.New("property not found")
	ErrPropertyNotString = errors.New("property is not a string")

	ErrNoNodes      = errors.New("no nodes available")
	ErrNodeNotFound = errors.New("node not found")
)

func (s ServiceNodeSlice) GetProperty(name string) (any, error) {
	if len(s) == 0 {
		return nil, ErrNoNodes
	}

	for _, svc := range s {
		if val, ok := svc.Properties[name]; ok {
			return val, nil
		}
	}

	return nil, ErrPropertyNotFound
}

func (s ServiceNodeSlice) GetPropertyString(name string) (string, error) {
	val, err := s.GetProperty(name)
	if err != nil {
		return "", err
	}

	if str, ok := val.(string); ok {
		return str, nil
	}

	return "", ErrPropertyNotString
}

type ServiceNodePicker func(ServiceNodeSlice) (ServiceNode, error)

func RandomNodePicker() ServiceNodePicker {
	return func(nodes ServiceNodeSlice) (ServiceNode, error) {
		if len(nodes) == 0 {
			return ServiceNode{}, ErrNoNodes
		}

		return nodes[rand.Intn(len(nodes))], nil
	}
}

func IndexNodePicker(index int) ServiceNodePicker {
	return func(nodes ServiceNodeSlice) (ServiceNode, error) {
		if index < 0 || index >= len(nodes) {
			return ServiceNode{}, ErrNoNodes
		}

		return nodes[index], nil
	}
}

func NameNodePicker(name string) ServiceNodePicker {
	return func(nodes ServiceNodeSlice) (ServiceNode, error) {
		for _, node := range nodes {
			if node.NodeName == name {
				return node, nil
			}
		}

		return ServiceNode{}, ErrNodeNotFound
	}
}

func PredicateNodePicker(predicate func(ServiceNode) bool) ServiceNodePicker {
	return func(nodes ServiceNodeSlice) (ServiceNode, error) {
		for _, node := range nodes {
			if predicate(node) {
				return node, nil
			}
		}

		return ServiceNode{}, ErrNodeNotFound
	}
}

func (s ServiceNodeSlice) PickNode(picker ServiceNodePicker) (ServiceNode, error) {
	return picker(s)
}
