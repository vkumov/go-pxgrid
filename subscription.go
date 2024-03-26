package gopxgrid

import (
	"context"
	"encoding/json"

	"github.com/go-stomp/stomp/v3"
)

type (
	Subscription[T any] struct {
		*stomp.Subscription

		C chan *Message[T]
	}

	Message[T any] struct {
		*stomp.Message

		Body           T
		UnmarshalError error
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

func translator[T any](in chan *stomp.Message) chan *Message[T] {
	out := make(chan *Message[T])

	go func() {
		defer close(out)

		for msg := range in {
			if msg.Err != nil {
				out <- &Message[T]{
					Message: msg,
				}
				continue
			}

			var body T
			if err := json.Unmarshal(msg.Body, &body); err != nil {
				out <- &Message[T]{
					Message: msg,
					Body:    body,
				}
			} else {
				out <- &Message[T]{
					Message:        msg,
					UnmarshalError: err,
				}
			}
		}
	}()

	return out
}

func subscribe[T any](ctx context.Context, s PubSubSubscriber, node ServiceNode, topic string) (*Subscription[T], error) {
	sub, err := s.Subscribe(ctx, topic, node)
	if err != nil {
		return nil, err
	}

	return &Subscription[T]{
		Subscription: sub,
		C:            translator[T](sub.C),
	}, nil
}
