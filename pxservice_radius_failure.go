package gopxgrid

import (
	"fmt"
)

type (
	Failure struct {
		ID                       string   `json:"id"`
		Timestamp                string   `json:"timestamp"`
		FailureReason            string   `json:"failureReason"`
		UserName                 string   `json:"userName"`
		ServerName               string   `json:"serverName"`
		CallingStationID         string   `json:"callingStationId"`
		AuditSessionID           string   `json:"auditSessionId"`
		NASIPAddress             string   `json:"nasIpAddress"`
		NASPortID                string   `json:"nasPortId"`
		NASPortType              string   `json:"nasPortType"`
		IPAddresses              []string `json:"ipAddresses"`
		MACAddress               string   `json:"macAddress"`
		MessageCode              int      `json:"messageCode"`
		DestinationIPAddress     string   `json:"destinationIpAddress"`
		UserType                 string   `json:"userType"`
		AccessService            string   `json:"accessService"`
		IdentityStore            string   `json:"identityStore"`
		IdentityGroup            string   `json:"identityGroup"`
		AuthenticationMethod     string   `json:"authenticationMethod"`
		AuthenticationProtocol   string   `json:"authenticationProtocol"`
		ServiceType              string   `json:"serviceType"`
		NetworkDeviceName        string   `json:"networkDeviceName"`
		DeviceType               string   `json:"deviceType"`
		Location                 string   `json:"location"`
		SelectedAznProfiles      string   `json:"selectedAznProfiles"`
		PostureStatus            string   `json:"postureStatus"`
		CTSSecurityGroup         string   `json:"ctsSecurityGroup"`
		Response                 string   `json:"response"`
		ResponseTime             int      `json:"responseTime"`
		ExecutionSteps           string   `json:"executionSteps"`
		CredentialCheck          string   `json:"credentialCheck"`
		EndpointProfile          string   `json:"endpointProfile"`
		MDMServerName            string   `json:"mdmServerName"`
		PolicySetName            string   `json:"policySetName"`
		AuthorizationRule        string   `json:"authorizationRule"`
		MSEResponseTime          string   `json:"mseResponseTime"`
		MSEServerName            string   `json:"mseServerName"`
		OriginalCallingStationID string   `json:"originalCallingStationId"`
	}

	FailureTopicMessage struct {
		Sequence int       `json:"sequence"`
		Failures []Failure `json:"failures"`
	}

	RadiusFailurePropsProvider interface {
		RestBaseURL() (string, error)
		WSPubsubService() (string, error)
		FailureTopic() (string, error)
	}

	RadiusFailureTopic string

	RadiusFailureSubscriber interface {
		OnFailureTopic() Subscriber[FailureTopicMessage]
	}

	RadiusFailureRest interface {
		GetFailures() CallFinalizer[*[]Failure]
		GetFailureByID(id string) CallFinalizer[*Failure]
	}

	RadiusFailure interface {
		PxGridService

		Rest() RadiusFailureRest

		RadiusFailureSubscriber

		Properties() RadiusFailurePropsProvider
	}

	pxGridRadiusFailure struct {
		pxGridService
	}
)

const (
	RadiusFailureTopicFailure RadiusFailureTopic = "failureTopic"

	RadiusFailureServiceName = "com.cisco.ise.radius"
)

func NewPxGridRadiusFailure(ctrl *PxGridConsumer) RadiusFailure {
	return &pxGridRadiusFailure{
		pxGridService{
			name: RadiusFailureServiceName,
			ctrl: ctrl,
			log:  ctrl.cfg.Logger.With("svc", RadiusFailureServiceName),
		},
	}
}

func (r *pxGridRadiusFailure) Rest() RadiusFailureRest {
	return r
}

// GetFailures retrieves the list of failures from the radius failure service
func (r *pxGridRadiusFailure) GetFailures() CallFinalizer[*[]Failure] {
	type response struct {
		Failures []Failure `json:"failures"`
	}

	return newCall[*[]Failure](
		&r.pxGridService,
		"getFailures",
		map[string]any{},
		func(r *Response) (*[]Failure, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			return &r.Result.(*response).Failures, nil
		},
	)
}

// GetFailureByID retrieves a failure by its ID
func (r *pxGridRadiusFailure) GetFailureByID(id string) CallFinalizer[*Failure] {
	if id == "" {
		return newFailedCall[*Failure](ErrInvalidInput)
	}

	return newCall[*Failure](
		&r.pxGridService,
		"getFailureById",
		map[string]any{"id": id},
		simpleResultMapper[*Failure],
	)
}

func (r *pxGridRadiusFailure) Properties() RadiusFailurePropsProvider {
	return r
}

func (r *pxGridRadiusFailure) RestBaseURL() (string, error) {
	return r.nodes.GetPropertyString("restBaseUrl")
}

func (r *pxGridRadiusFailure) WSPubsubService() (string, error) {
	return r.nodes.GetPropertyString("wsPubsubService")
}

func (r *pxGridRadiusFailure) FailureTopic() (string, error) {
	return r.nodes.GetPropertyString(string(RadiusFailureTopicFailure))
}

func (r *pxGridRadiusFailure) OnFailureTopic() Subscriber[FailureTopicMessage] {
	return newSubscriber[FailureTopicMessage](
		&r.pxGridService,
		string(RadiusFailureTopicFailure),
		r.WSPubsubService,
	)
}
