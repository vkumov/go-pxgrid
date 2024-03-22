package gopxgrid

import "fmt"

type (
	Profile struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		FullName string `json:"fullName"`
	}

	ProfilerConfiguration interface {
		PxGridService

		GetProfiles() CallFinalizer[*[]Profile]
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
