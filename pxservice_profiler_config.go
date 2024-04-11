package gopxgrid

import (
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

	ProfilerConfigurationTopic string

	ProfilerConfigurationSubscriber interface {
		OnProfileTopic() Subscriber[ProfilerTopicMessage]
	}

	ProfilerConfiguration interface {
		PxGridService

		GetProfiles() CallFinalizer[*[]Profile]

		ProfilerConfigurationSubscriber

		Properties() ProfilerConfigurationPropsProvider
	}

	pxGridProfilerConfiguration struct {
		pxGridService
	}
)

const (
	ProfilerConfigurationTopicProfile ProfilerConfigurationTopic = "topic"

	ProfilerConfigurationServiceName = "com.cisco.ise.config.profiler"
)

func NewPxGridProfilerConfiguration(ctrl *PxGridConsumer) ProfilerConfiguration {
	return &pxGridProfilerConfiguration{
		pxGridService{
			name: ProfilerConfigurationServiceName,
			ctrl: ctrl,
			log:  ctrl.cfg.Logger.With("svc", ProfilerConfigurationServiceName),
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
	return s.nodes.GetPropertyString(string(ProfilerConfigurationTopicProfile))
}

func (s *pxGridProfilerConfiguration) OnProfileTopic() Subscriber[ProfilerTopicMessage] {
	return newSubscriber[ProfilerTopicMessage](
		&s.pxGridService,
		string(ProfilerConfigurationTopicProfile),
		s.WSPubsubService,
	)
}
