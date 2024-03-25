package gopxgrid

import (
	"errors"
	"fmt"
)

type ANCPolicyAction string

const (
	ANCActionQuarantine     ANCPolicyAction = "QUARANTINE"
	ANCActionShutdown       ANCPolicyAction = "SHUT_DOWN"
	ANCActionPortBounce     ANCPolicyAction = "PORT_BOUNCE"
	ANCActionReAuthenticate ANCPolicyAction = "RE_AUTHENTICATE"
)

func (a ANCPolicyAction) Valid() bool {
	switch a {
	case ANCActionQuarantine, ANCActionShutdown, ANCActionPortBounce, ANCActionReAuthenticate:
		return true
	}
	return false
}

type (
	ANCPolicy struct {
		Name    string            `json:"name"`
		Actions []ANCPolicyAction `json:"actions"`
		ID      string            `json:"id"`
	}

	ANCStatus string

	ANCOperationStatus struct {
		ID            string    `json:"operationId"`
		Status        ANCStatus `json:"status"`
		MACAddress    string    `json:"macAddress,omitempty"`
		NASIPAddress  string    `json:"nasIpAddress,omitempty"`
		FailureReason string    `json:"failureReason,omitempty"`
	}

	ANCEndpoint struct {
		MACAddress string `json:"macAddress"`
		PolicyName string `json:"policyName"`
		ID         string `json:"id"`
	}

	ANCApplyPolicyRequest struct {
		Policy       string `json:"policy"`
		MACAddress   string `json:"macAddress"`
		NASIPAddress string `json:"nasIpAddress"`
		SessionID    string `json:"sessionId,omitempty"`
		NASPortID    string `json:"nasPortId,omitempty"`
		IPAddress    string `json:"ipAddress,omitempty"`
		UserName     string `json:"userName,omitempty"`
	}

	ANCClearPolicyRequest struct {
		MACAddress   string `json:"macAddress"`
		NASIPAddress string `json:"nasIpAddress"`
	}
)

const (
	ANCStatusSuccess ANCStatus = "SUCCESS"
	ANCStatusFailure ANCStatus = "FAILURE"
	ANCStatusRunning ANCStatus = "RUNNING"
)

type ANCConfigPropsProvider interface {
	RestBaseURL() (string, error)
	WSPubsubService() (string, error)
	StatusTopic() (string, error)
}

type ANCConfig interface {
	PxGridService

	GetPolicies() CallFinalizer[*[]ANCPolicy]
	GetPolicyByName(name string) CallFinalizer[*ANCPolicy]
	CreatePolicy(policy ANCPolicy) NoResultCallFinalizer
	DeletePolicyByName(name string) NoResultCallFinalizer

	GetEndpoints() CallFinalizer[*[]ANCEndpoint]
	GetEndpointPolicies() CallFinalizer[*[]ANCEndpoint]
	GetEndpointByMAC(mac string) CallFinalizer[*ANCEndpoint]
	GetEndpointByNasIPAddress(mac, nasIP string) CallFinalizer[*ANCEndpoint]
	ApplyEndpointByIPAddress(ip, policyName string) CallFinalizer[*ANCOperationStatus]
	ApplyEndpointByMACAddress(mac, policyName string) CallFinalizer[*ANCOperationStatus]
	ClearEndpointByIPAddress(ip, policyName string) CallFinalizer[*ANCOperationStatus]
	ClearEndpointByMACAddress(mac, policyName string) CallFinalizer[*ANCOperationStatus]

	ApplyEndpointPolicy(request ANCApplyPolicyRequest) CallFinalizer[*ANCOperationStatus]
	ClearEndpointPolicy(request ANCClearPolicyRequest) CallFinalizer[*ANCOperationStatus]

	GetOperationStatus(operationID string) CallFinalizer[*ANCOperationStatus]

	SubscribeStatusTopic() (*Subscription[ANCOperationStatus], error)

	Properties() ANCConfigPropsProvider
}

type pxGridANC struct {
	pxGridService
}

func NewPxGridANCConfig(ctrl Controller) ANCConfig {
	return &pxGridANC{
		pxGridService: pxGridService{
			name: "com.cisco.ise.config.anc",
			ctrl: ctrl,
		},
	}
}

func (a *pxGridANC) GetPolicies() CallFinalizer[*[]ANCPolicy] {
	type response struct {
		Policies []ANCPolicy `json:"policies"`
	}

	return newCall[*[]ANCPolicy](
		&a.pxGridService,
		"getPolicies",
		map[string]any{},
		func(r *Response) (*[]ANCPolicy, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			return &r.Result.(*response).Policies, nil
		},
	)
}

func (a *pxGridANC) GetPolicyByName(name string) CallFinalizer[*ANCPolicy] {
	if name == "" {
		return newFailedCall[*ANCPolicy](ErrInvalidInput)
	}

	return newCall[*ANCPolicy](
		&a.pxGridService,
		"getPolicyByName",
		map[string]any{"name": name},
		simpleResultMapper[*ANCPolicy],
	)
}

func (a *pxGridANC) CreatePolicy(policy ANCPolicy) NoResultCallFinalizer {
	if policy.Name == "" || len(policy.Actions) == 0 {
		return newFailedNoResultCall(ErrInvalidInput)
	}

	payload := map[string]any{
		"name":    policy.Name,
		"actions": policy.Actions,
	}

	return newNoResultCall[any](
		&a.pxGridService,
		"createPolicy",
		payload,
		simpleNoResultMapper,
	)
}

func (a *pxGridANC) DeletePolicyByName(name string) NoResultCallFinalizer {
	if name == "" {
		return newFailedNoResultCall(ErrInvalidInput)
	}

	return newNoResultCall[any](
		&a.pxGridService,
		"deletePolicyByName",
		map[string]any{"name": name},
		simpleNoResultMapper,
	)
}

func (a *pxGridANC) GetEndpoints() CallFinalizer[*[]ANCEndpoint] {
	type response struct {
		Endpoints []ANCEndpoint `json:"endpoints"`
	}

	return newCall[*[]ANCEndpoint](
		&a.pxGridService,
		"getEndpoints",
		map[string]any{},
		func(r *Response) (*[]ANCEndpoint, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			return &r.Result.(*response).Endpoints, nil
		},
	)
}

func (a *pxGridANC) GetEndpointPolicies() CallFinalizer[*[]ANCEndpoint] {
	type response struct {
		Endpoints []ANCEndpoint `json:"endpoints"`
	}

	return newCall[*[]ANCEndpoint](
		&a.pxGridService,
		"getEndpointPolicies",
		map[string]any{},
		func(r *Response) (*[]ANCEndpoint, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			return &r.Result.(*response).Endpoints, nil
		},
	)
}

func (a *pxGridANC) GetEndpointByMAC(mac string) CallFinalizer[*ANCEndpoint] {
	if mac == "" {
		return newFailedCall[*ANCEndpoint](ErrInvalidInput)
	}

	return newCall[*ANCEndpoint](
		&a.pxGridService,
		"getEndpointByMAC",
		map[string]any{"macAddress": mac},
		simpleResultMapper[*ANCEndpoint],
	)
}

func (a *pxGridANC) GetEndpointByNasIPAddress(mac, nasIP string) CallFinalizer[*ANCEndpoint] {
	if mac == "" || nasIP == "" {
		return newFailedCall[*ANCEndpoint](ErrInvalidInput)
	}

	payload := map[string]any{
		"macAddress":   mac,
		"nasIpAddress": nasIP,
	}

	return newCall[*ANCEndpoint](
		&a.pxGridService,
		"getEndpointByNasIpAddress",
		payload,
		simpleResultMapper[*ANCEndpoint],
	)
}

func (a *pxGridANC) ApplyEndpointByIPAddress(ip, policyName string) CallFinalizer[*ANCOperationStatus] {
	if ip == "" || policyName == "" {
		return newFailedCall[*ANCOperationStatus](ErrInvalidInput)
	}

	payload := map[string]any{
		"ipAddress":  ip,
		"policyName": policyName,
	}

	return newCall[*ANCOperationStatus](
		&a.pxGridService,
		"applyEndpointByIpAddress",
		payload,
		simpleResultMapper[*ANCOperationStatus],
	)
}

func (a *pxGridANC) ApplyEndpointByMACAddress(mac, policyName string) CallFinalizer[*ANCOperationStatus] {
	if mac == "" || policyName == "" {
		return newFailedCall[*ANCOperationStatus](ErrInvalidInput)
	}

	payload := map[string]any{
		"macAddress": mac,
		"policyName": policyName,
	}

	return newCall[*ANCOperationStatus](
		&a.pxGridService,
		"applyEndpointByMacAddress",
		payload,
		simpleResultMapper[*ANCOperationStatus],
	)
}

func (a *pxGridANC) ClearEndpointByIPAddress(ip, policyName string) CallFinalizer[*ANCOperationStatus] {
	if ip == "" || policyName == "" {
		return newFailedCall[*ANCOperationStatus](ErrInvalidInput)
	}

	payload := map[string]any{
		"ipAddress":  ip,
		"policyName": policyName,
	}

	return newCall[*ANCOperationStatus](
		&a.pxGridService,
		"clearEndpointByIpAddress",
		payload,
		simpleResultMapper[*ANCOperationStatus],
	)
}

func (a *pxGridANC) ClearEndpointByMACAddress(mac, policyName string) CallFinalizer[*ANCOperationStatus] {
	if mac == "" || policyName == "" {
		return newFailedCall[*ANCOperationStatus](ErrInvalidInput)
	}

	payload := map[string]any{
		"macAddress": mac,
		"policyName": policyName,
	}

	return newCall[*ANCOperationStatus](
		&a.pxGridService,
		"clearEndpointByMacAddress",
		payload,
		simpleResultMapper[*ANCOperationStatus],
	)
}

func (a *pxGridANC) GetOperationStatus(operationID string) CallFinalizer[*ANCOperationStatus] {
	if operationID == "" {
		return newFailedCall[*ANCOperationStatus](ErrInvalidInput)
	}

	return newCall[*ANCOperationStatus](
		&a.pxGridService,
		"getOperationStatus",
		map[string]any{"operationId": operationID},
		simpleResultMapper[*ANCOperationStatus],
	)
}

func (a *pxGridANC) ApplyEndpointPolicy(request ANCApplyPolicyRequest) CallFinalizer[*ANCOperationStatus] {
	if request.MACAddress == "" || request.Policy == "" || request.NASIPAddress == "" {
		return newFailedCall[*ANCOperationStatus](ErrInvalidInput)
	}

	return newCall[*ANCOperationStatus](
		&a.pxGridService,
		"applyEndpointPolicy",
		request,
		simpleResultMapper[*ANCOperationStatus],
	)
}

func (a *pxGridANC) ClearEndpointPolicy(request ANCClearPolicyRequest) CallFinalizer[*ANCOperationStatus] {
	if request.MACAddress == "" || request.NASIPAddress == "" {
		return newFailedCall[*ANCOperationStatus](ErrInvalidInput)
	}

	return newCall[*ANCOperationStatus](
		&a.pxGridService,
		"clearEndpointPolicy",
		request,
		simpleResultMapper[*ANCOperationStatus],
	)
}

func (a *pxGridANC) Properties() ANCConfigPropsProvider {
	return a
}

func (a *pxGridANC) RestBaseURL() (string, error) {
	return a.nodes.GetPropertyString("restBaseUrl")
}

func (a *pxGridANC) WSPubsubService() (string, error) {
	return a.nodes.GetPropertyString("wsPubsubService")
}

func (a *pxGridANC) StatusTopic() (string, error) {
	return a.nodes.GetPropertyString("statusTopic")
}

func (a *pxGridANC) SubscribeStatusTopic() (*Subscription[ANCOperationStatus], error) {
	//
	return nil, errors.New("not implemented")
}
