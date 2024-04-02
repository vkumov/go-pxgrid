package gopxgrid

import (
	"fmt"
)

type (
	MDMEndpoint struct {
		MACAddress    string `json:"macAddress"`
		OSVersion     string `json:"osVersion"`
		Registered    bool   `json:"registered"`
		Compliant     bool   `json:"compliant"`
		DiskEncrypted bool   `json:"diskEncrypted"`
		JailBroken    bool   `json:"jailBroken"`
		PinLocked     bool   `json:"pinLocked"`
		Model         string `json:"model"`
		Manufacturer  string `json:"manufacturer"`
		IMEI          string `json:"imei"`
		MEID          string `json:"meid"`
		UDID          string `json:"udid"`
		SerialNumber  string `json:"serialNumber"`
		Location      string `json:"location"`
		DeviceManager string `json:"deviceManager"`
		LastSyncTime  string `json:"lastSyncTime"`
	}

	MDMEndpointType string

	MDMOSType string

	MDMPropsProvider interface {
		RestBaseURL() (string, error)
		WSPubsubService() (string, error)
		EndpointTopic() (string, error)
	}

	MDMTopic string

	MDMSubscriber interface {
		OnEndpointTopic() Subscriber[MDMEndpoint]
	}

	MDM interface {
		PxGridService

		GetEndpoints(filter *MDMEndpoint) CallFinalizer[*[]MDMEndpoint]
		GetEndpointByMacAddress(macAddress string) CallFinalizer[*MDMEndpoint]
		GetEndpointsByType(endpointType MDMEndpointType) CallFinalizer[*[]MDMEndpoint]
		GetEndpointsByOsType(osType MDMOSType) CallFinalizer[*[]MDMEndpoint]

		MDMSubscriber

		Properties() MDMPropsProvider
	}

	pxGridMDM struct {
		pxGridService
	}
)

const (
	MDMEndpointTypeNonCompliant MDMEndpointType = "NON_COMPLIANT"
	MDMEndpointTypeRegistered   MDMEndpointType = "REGISTERED"
	MDMEndpointTypeDisconnected MDMEndpointType = "DISCONNECTED"

	MDMOSTypeAndroid MDMOSType = "ANDROID"
	MDMOSTypeIOS     MDMOSType = "IOS"
	MDMOSTypeWindows MDMOSType = "WINDOWS"

	MDMTopicEndpoint MDMTopic = "endpointTopic"
)

func NewPxGridMDM(ctrl *PxGridConsumer) MDM {
	return &pxGridMDM{
		pxGridService{
			name: "com.cisco.ise.mdm",
			ctrl: ctrl,
			log:  ctrl.cfg.Logger.With("svc", "com.cisco.ise.mdm"),
		},
	}
}

// GetEndpoints retrieves the endpoints from the MDM service
func (s *pxGridMDM) GetEndpoints(filter *MDMEndpoint) CallFinalizer[*[]MDMEndpoint] {
	payload := map[string]any{}
	if filter != nil {
		payload["filter"] = filter
	}

	type response struct {
		Endpoints []MDMEndpoint `json:"endpoints"`
	}

	return newCall[*[]MDMEndpoint](
		&s.pxGridService,
		"getEndpoints",
		payload,
		func(r *Response) (*[]MDMEndpoint, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			return &r.Result.(*response).Endpoints, nil
		},
	)
}

// GetEndpointByMacAddress retrieves an endpoint by its MAC address
func (s *pxGridMDM) GetEndpointByMacAddress(macAddress string) CallFinalizer[*MDMEndpoint] {
	if macAddress == "" {
		return newFailedCall[*MDMEndpoint](ErrInvalidInput)
	}

	return newCall[*MDMEndpoint](
		&s.pxGridService,
		"getEndpointByMacAddress",
		map[string]any{"macAddress": macAddress},
		simpleResultMapper[*MDMEndpoint],
	)
}

// GetEndpointsByType retrieves the endpoints by type
func (s *pxGridMDM) GetEndpointsByType(endpointType MDMEndpointType) CallFinalizer[*[]MDMEndpoint] {
	payload := map[string]any{"type": endpointType}

	type response struct {
		Endpoints []MDMEndpoint `json:"endpoints"`
	}

	return newCall[*[]MDMEndpoint](
		&s.pxGridService,
		"getEndpointsByType",
		payload,
		func(r *Response) (*[]MDMEndpoint, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			return &r.Result.(*response).Endpoints, nil
		},
	)
}

// GetEndpointsByOsType retrieves the endpoints by OS type
func (s *pxGridMDM) GetEndpointsByOsType(osType MDMOSType) CallFinalizer[*[]MDMEndpoint] {
	payload := map[string]any{"osType": osType}

	type response struct {
		Endpoints []MDMEndpoint `json:"endpoints"`
	}

	return newCall[*[]MDMEndpoint](
		&s.pxGridService,
		"getEndpointsByOsType",
		payload,
		func(r *Response) (*[]MDMEndpoint, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			return &r.Result.(*response).Endpoints, nil
		},
	)
}

func (s *pxGridMDM) Properties() MDMPropsProvider {
	return s
}

func (s *pxGridMDM) RestBaseURL() (string, error) {
	return s.nodes.GetPropertyString("restBaseUrl")
}

func (s *pxGridMDM) WSPubsubService() (string, error) {
	return s.nodes.GetPropertyString("wsPubsubService")
}

func (s *pxGridMDM) EndpointTopic() (string, error) {
	return s.nodes.GetPropertyString(string(MDMTopicEndpoint))
}

func (s *pxGridMDM) OnEndpointTopic() Subscriber[MDMEndpoint] {
	return newSubscriber[MDMEndpoint](
		&s.pxGridService,
		string(MDMTopicEndpoint),
		s.WSPubsubService,
	)
}
