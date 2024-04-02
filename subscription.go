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

type Subscriber[T any] interface {
	WithServiceNodePicker(picker ServiceNodePickerFactory) Subscriber[T]
	WithPubSubNodePicker(picker ServiceNodePickerFactory) Subscriber[T]
	WithExplicitPubSub(pubsub PubSub) Subscriber[T]
	Subscribe(ctx context.Context) (*Subscription[T], error)
}

type subscriber[T any] struct {
	svc           *pxGridService
	topicProperty string
	pubsub        PubSub
	pubsubGetter  func() (string, error)

	svcNodePicker    ServiceNodePickerFactory
	pubSubNodePicker ServiceNodePickerFactory
}

func newSubscriber[T any](svc *pxGridService, topic string, pubsubGetter func() (string, error)) Subscriber[T] {
	return &subscriber[T]{
		svc:           svc,
		topicProperty: topic,
		pubsubGetter:  pubsubGetter,
	}
}

func (s *subscriber[T]) WithExplicitPubSub(pubsub PubSub) Subscriber[T] {
	s.pubsub = pubsub
	return s
}

func (s *subscriber[T]) WithServiceNodePicker(picker ServiceNodePickerFactory) Subscriber[T] {
	s.svcNodePicker = picker
	return s
}

func (s *subscriber[T]) WithPubSubNodePicker(picker ServiceNodePickerFactory) Subscriber[T] {
	s.pubSubNodePicker = picker
	return s
}

func (s *subscriber[T]) getPubSubServiceName(ctx context.Context) (string, error) {
	if s.pubsubGetter != nil {
		return s.pubsubGetter()
	}

	psNameRaw, err := s.svc.FindProperty(ctx, "wsPubsubService", s.svcNodePicker)
	if err != nil {
		return "", err
	}

	pubSubServiceName, ok := psNameRaw.(string)
	if !ok || pubSubServiceName == "" {
		return "", ErrPropertyNotFound
	}

	return pubSubServiceName, nil
}

func (s *subscriber[T]) getTopic(ctx context.Context) (string, error) {
	topicRaw, err := s.svc.FindProperty(ctx, s.topicProperty, s.svcNodePicker)
	if err != nil {
		return "", err
	}

	topic, ok := topicRaw.(string)
	if !ok || topic == "" {
		return "", ErrPropertyNotFound
	}

	return topic, nil
}

func (s *subscriber[T]) populatePubSub(ctx context.Context) error {
	if s.pubsub != nil {
		return nil
	}

	s.svc.log.Debug("Populating PubSub in subscriber")
	pubSubServiceName, err := s.getPubSubServiceName(ctx)
	if err != nil {
		return err
	}

	s.pubsub = s.svc.ctrl.PubSub(pubSubServiceName)
	return nil
}

func (s *subscriber[T]) Subscribe(ctx context.Context) (*Subscription[T], error) {
	if s.svcNodePicker == nil {
		s.svcNodePicker = OrderedNodePicker()
	}

	s.svc.log.Debug("Subscribing to topic", "topicProperty", s.topicProperty)
	if err := s.populatePubSub(ctx); err != nil {
		return nil, err
	}

	if err := s.pubsub.CheckNodes(ctx); err != nil {
		return nil, err
	}

	topic, err := s.getTopic(ctx)
	if err != nil {
		return nil, err
	}
	s.svc.log.Debug("Subscribing to topic", "topic", topic)

	sub, err := s.pubsub.Subscribe(ctx, s.pubSubNodePicker, topic)
	if err != nil {
		return nil, err
	}

	return &Subscription[T]{
		Subscription: sub,
		C:            translator[T](sub.C),
	}, nil
}
