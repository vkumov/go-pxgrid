package gopxgrid

import (
	"encoding/json"

	"github.com/go-stomp/stomp/v3"
)

type (
	Subscription[T any] struct {
		*stomp.Subscription

		C chan T
	}
)

func (s *Subscription[T]) Read() (T, error) {
	msg, err := s.Subscription.Read()
	if err != nil {
		var zero T
		return zero, err
	}

	var result T
	if err := json.Unmarshal(msg.Body, &result); err != nil {
		var zero T
		return zero, err
	}

	return result, nil
}
