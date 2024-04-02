package gopxgrid

import (
	"errors"
	"math/rand"
	"slices"
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

	ErrNoNodePicked = errors.New("no node picked")
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

type ServiceNodePicker interface {
	PickNode() (ServiceNode, bool, error)
}

type ServiceNodePickerFactory func(ServiceNodeSlice) ServiceNodePicker

type predicateNodePicker struct {
	predicate func(int, ServiceNode) bool
	nodes     ServiceNodeSlice
	last      int
}

func (p *predicateNodePicker) PickNode() (ServiceNode, bool, error) {
	if len(p.nodes) == 0 {
		return ServiceNode{}, false, ErrNoNodes
	}

	for i := p.last; i < len(p.nodes); i++ {
		if p.predicate(i, p.nodes[i]) {
			p.last = i
			return p.nodes[i], i < len(p.nodes)-1, nil
		}
	}

	return ServiceNode{}, false, ErrNodeNotFound
}

type randomNodePicker struct {
	nodes       ServiceNodeSlice
	indexesLeft []int
}

func (p *randomNodePicker) PickNode() (ServiceNode, bool, error) {
	if len(p.indexesLeft) == 0 {
		return ServiceNode{}, false, ErrNoNodes
	}

	index := rand.Intn(len(p.indexesLeft))
	node := p.nodes[p.indexesLeft[index]]
	p.indexesLeft = append(p.indexesLeft[:index], p.indexesLeft[index+1:]...)

	return node, len(p.indexesLeft) > 0, nil
}

func RandomNodePicker() ServiceNodePickerFactory {
	return func(nodes ServiceNodeSlice) ServiceNodePicker {
		indexes := make([]int, len(nodes))
		for i := range indexes {
			indexes[i] = i
		}

		return &randomNodePicker{
			nodes:       nodes,
			indexesLeft: indexes,
		}
	}
}

func OrderedNodePicker() ServiceNodePickerFactory {
	return func(nodes ServiceNodeSlice) ServiceNodePicker {
		return &predicateNodePicker{
			predicate: func(int, ServiceNode) bool {
				return true
			},
			nodes: nodes,
		}
	}
}

func IndexNodePicker(index ...int) ServiceNodePickerFactory {
	return func(nodes ServiceNodeSlice) ServiceNodePicker {
		return &predicateNodePicker{
			predicate: func(i int, _ ServiceNode) bool {
				return slices.Contains(index, i)
			},
			nodes: nodes,
		}
	}
}

func NameNodePicker(name ...string) ServiceNodePickerFactory {
	return func(nodes ServiceNodeSlice) ServiceNodePicker {
		return &predicateNodePicker{
			predicate: func(_ int, node ServiceNode) bool {
				return slices.Contains(name, node.Name)
			},
			nodes: nodes,
		}
	}
}

func PredicateNodePicker(predicate func(ServiceNode) bool) ServiceNodePickerFactory {
	return func(nodes ServiceNodeSlice) ServiceNodePicker {
		return &predicateNodePicker{
			predicate: func(_ int, node ServiceNode) bool {
				return predicate(node)
			},
			nodes: nodes,
		}
	}
}
