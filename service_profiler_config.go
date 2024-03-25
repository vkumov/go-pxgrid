package gopxgrid

import (
	"errors"
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
		OnProfileTopic() (*Subscription[ProfilerTopicMessage], error)
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

func NewPxGridProfilerConfiguration(ctrl Controller) ProfilerConfiguration {
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

func (s *pxGridProfilerConfiguration) OnProfileTopic() (*Subscription[ProfilerTopicMessage], error) {
	return nil, errors.New("not implemented")
}
