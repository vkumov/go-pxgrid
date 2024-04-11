package gopxgrid

import "fmt"

type (
	SysHealth struct {
		Timestamp       string  `json:"timestamp"`
		ServerName      string  `json:"serverName"`
		IOWait          float64 `json:"ioWait"`
		CPUUsage        float64 `json:"cpuUsage"`
		MemoryUsage     float64 `json:"memoryUsage"`
		DiskUsageRoot   float64 `json:"diskUsageRoot"`
		DiskUsageOpt    float64 `json:"diskUsageOpt"`
		LoadAverage     float64 `json:"loadAverage"`
		NetworkSent     float64 `json:"networkSent"`
		NetworkReceived float64 `json:"networkReceived"`
	}

	SysPerformance struct {
		Timestamp     string  `json:"timestamp"`
		ServerName    string  `json:"serverName"`
		RADIUSRate    float64 `json:"radiusRate"`
		RADIUSCount   float64 `json:"radiusCount"`
		RADIUSLatency float64 `json:"radiusLatency"`
	}

	SystemHealthPropsProvider interface {
		RestBaseURL() (string, error)
	}

	SystemHealth interface {
		PxGridService

		GetHealths(nodeName string, startTimestamp string) CallFinalizer[*[]SysHealth]
		GetPerformances(nodeName string, startTimestamp string) CallFinalizer[*[]SysPerformance]

		Properties() SystemHealthPropsProvider
	}

	pxGridSystemHealth struct {
		pxGridService
	}
)

const (
	SystemHealthServiceName = "com.cisco.ise.system"
)

func NewPxGridSystemHealth(ctrl *PxGridConsumer) SystemHealth {
	return &pxGridSystemHealth{
		pxGridService{
			name: SystemHealthServiceName,
			ctrl: ctrl,
			log:  ctrl.cfg.Logger.With("svc", SystemHealthServiceName),
		},
	}
}

func (s *pxGridSystemHealth) GetHealths(nodeName string, startTimestamp string) CallFinalizer[*[]SysHealth] {
	payload := map[string]any{}
	if nodeName != "" {
		payload["nodeName"] = nodeName
	}
	if startTimestamp != "" {
		payload["startTimestamp"] = startTimestamp
	}

	type response struct {
		Healths []SysHealth `json:"healths"`
	}

	return newCall[*[]SysHealth](
		&s.pxGridService,
		"getHealths",
		payload,
		func(r *Response) (*[]SysHealth, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			return &r.Result.(*response).Healths, nil
		},
	)
}

func (s *pxGridSystemHealth) GetPerformances(nodeName string, startTimestamp string) CallFinalizer[*[]SysPerformance] {
	payload := map[string]any{}
	if nodeName != "" {
		payload["nodeName"] = nodeName
	}
	if startTimestamp != "" {
		payload["startTimestamp"] = startTimestamp
	}

	type response struct {
		Performances []SysPerformance `json:"performances"`
	}

	return newCall[*[]SysPerformance](
		&s.pxGridService,
		"getPerformances",
		payload,
		func(r *Response) (*[]SysPerformance, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			return &r.Result.(*response).Performances, nil
		},
	)
}

func (s *pxGridSystemHealth) Properties() SystemHealthPropsProvider {
	return s
}

func (s *pxGridSystemHealth) RestBaseURL() (string, error) {
	return s.nodes.GetPropertyString("restBaseUrl")
}
