package gopxgrid

import (
	"context"
	"fmt"
)

type (
	Profile struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		FullName string `json:"fullName"`
	}

	ProfilerTopicMessage struct {
		OperationType OperationType `json:"operation"`
		Profile       Profile       `json:"profile"`
	}

	ProfilerConfigurationPropsProvider interface {
		RestBaseURL() (string, error)
		WSPubsubService() (string, error)
		Topic() (string, error)
	}

	ProfilerConfigurationSubscriber interface {
		OnProfileTopic(ctx context.Context, nodePicker ServiceNodePicker) (*Subscription[ProfilerTopicMessage], error)
	}

	ProfilerConfiguration interface {
		PxGridService

		GetProfiles() CallFinalizer[*[]Profile]

		Subscribe() ProfilerConfigurationSubscriber

		Properties() ProfilerConfigurationPropsProvider
	}

	pxGridProfilerConfiguration struct {
		pxGridService
	}
)

func NewPxGridProfilerConfiguration(ctrl *PxGridConsumer) ProfilerConfiguration {
	return &pxGridProfilerConfiguration{
		pxGridService{
			name: "com.cisco.ise.config.profiler",
			ctrl: ctrl,
		},
	}
}

// GetProfiles retrieves the list of profiles from the profiler configuration service
func (s *pxGridProfilerConfiguration) GetProfiles() CallFinalizer[*[]Profile] {
	type response struct {
		Profiles []Profile `json:"profiles"`
	}

	return newCall[*[]Profile](
		&s.pxGridService,
		"getProfiles",
		map[string]any{},
		func(r *Response) (*[]Profile, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			return &r.Result.(*response).Profiles, nil
		},
	)
}

func (s *pxGridProfilerConfiguration) Properties() ProfilerConfigurationPropsProvider {
	return s
}

func (s *pxGridProfilerConfiguration) RestBaseURL() (string, error) {
	return s.nodes.GetPropertyString("restBaseUrl")
}

func (s *pxGridProfilerConfiguration) WSPubsubService() (string, error) {
	return s.nodes.GetPropertyString("wsPubsubService")
}

func (s *pxGridProfilerConfiguration) Topic() (string, error) {
	return s.nodes.GetPropertyString("topic")
}

func (s *pxGridProfilerConfiguration) Subscribe() ProfilerConfigurationSubscriber {
	return s
}

func (s *pxGridProfilerConfiguration) OnProfileTopic(ctx context.Context, nodePicker ServiceNodePicker) (*Subscription[ProfilerTopicMessage], error) {
	node, err := nodePicker(s.nodes)
	if err != nil {
		return nil, err
	}

	topic, err := s.Topic()
	if err != nil {
		return nil, err
	}

	return subscribe[ProfilerTopicMessage](ctx, s.ctrl.PubSub(), node, topic)
}
