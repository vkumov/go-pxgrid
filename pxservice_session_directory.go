package gopxgrid

import (
	"fmt"
)

type (
	SessionState string

	Session struct {
		Timestamp                string       `json:"timestamp"`
		State                    SessionState `json:"state"`
		MacAddress               string       `json:"macAddress"`
		IPAddresses              []string     `json:"ipAddresses"`
		CallingStationID         string       `json:"callingStationId"`
		CalledStationID          string       `json:"calledStationId"`
		AuditSessionID           string       `json:"auditSessionId"`
		UserName                 string       `json:"userName"`
		NasIPAddress             string       `json:"nasIpAddress"`
		NasPortID                string       `json:"nasPortId"`
		NasPortType              string       `json:"nasPortType"`
		NasIdentifier            string       `json:"nasIdentifier"`
		SelectedAuthzProfiles    []string     `json:"selectedAuthzProfiles"`
		PostureStatus            string       `json:"postureStatus"`
		EndpointProfile          string       `json:"endpointProfile"`
		EndpointOperatingSystem  string       `json:"endpointOperatingSystem"`
		CTSSecurityGroup         string       `json:"ctsSecurityGroup"`
		ADNormalizedUser         string       `json:"adNormalizedUser"`
		ADUserDomainName         string       `json:"adUserDomainName"`
		ADHostDomainName         string       `json:"adHostDomainName"`
		ADUserNetBiosName        string       `json:"adUserNetBiosName"`
		ADHostNetBiosName        string       `json:"adHostNetBiosName"`
		ADUserResolvedIdentities string       `json:"adUserResolvedIdentities"`
		ADUserResolvedDNS        string       `json:"adUserResolvedDns"`
		ADHostResolvedIdentities string       `json:"adHostResolvedIdentities"`
		ADHostResolvedDNS        string       `json:"adHostResolvedDns"`
		ADUserSamAccountName     string       `json:"adUserSamAccountName"`
		ADHostSamAccountName     string       `json:"adHostSamAccountName"`
		ADUserQualifiedName      string       `json:"adUserQualifiedName"`
		ADHostQualifiedName      string       `json:"adHostQualifiedName"`
		Providers                []string     `json:"providers"`
		EndpointCheckResult      string       `json:"endpointCheckResult"`
		EndpointCheckTime        string       `json:"endpointCheckTime"`
		IdentitySourcePortStart  string       `json:"identitySourcePortStart"`
		IdentitySourcePortEnd    string       `json:"identitySourcePortEnd"`
		IdentitySourcePortFirst  string       `json:"identitySourcePortFirst"`
		TerminalServerAgentID    string       `json:"terminalServerAgentId"`
		IsMachineAuthentication  string       `json:"isMachineAuthentication"`
		ServiceType              string       `json:"serviceType"`
		TunnelPrivateGroupID     string       `json:"tunnelPrivateGroupId"`
		AirespaceWlanID          string       `json:"airespaceWlanId"`
		NetworkDeviceProfileName string       `json:"networkDeviceProfileName"`
		RadiusFlowType           string       `json:"radiusFlowType"`
		SSID                     string       `json:"ssid"`
		ANCPolicy                string       `json:"ancPolicy"`
		MDMMacAddress            string       `json:"mdmMacAddress"`
		MDMOSVersion             string       `json:"mdmOsVersion"`
		MDMRegistered            bool         `json:"mdmRegistered"`
		MDMCompliant             bool         `json:"mdmCompliant"`
		MDMDiskEncrypted         bool         `json:"mdmDiskEncrypted"`
		MDMJailBroken            bool         `json:"mdmJailBroken"`
		MDMPinLocked             bool         `json:"mdmPinLocked"`
		MDMModel                 string       `json:"mdmModel"`
		MDMManufacturer          string       `json:"mdmManufacturer"`
		MDMIMEI                  string       `json:"mdmImei"`
		MDMMEID                  string       `json:"mdmMeid"`
		MDMUDID                  string       `json:"mdmUdid"`
		MDMSerialNumber          string       `json:"mdmSerialNumber"`
		MDMLocation              string       `json:"mdmLocation"`
		MDMDeviceManager         string       `json:"mdmDeviceManager"`
		MDMLastSyncTime          string       `json:"mdmLastSyncTime"`
		VirtualNetwork           string       `json:"virtualNetwork"`
	}

	GroupType string

	Group struct {
		Name string    `json:"name"`
		Type GroupType `json:"type"`
	}

	SessionTopicMessage struct {
		Sequence int       `json:"sequence"`
		Sessions []Session `json:"sessions"`
	}

	GroupTopicMessage struct {
		UserGroups []Group `json:"userGroups"`
	}

	SessionDirectoryPropsProvider interface {
		RestBaseURL() (string, error)
		WSPubsubService() (string, error)
		SessionTopic() (string, error)
		SessionTopicAll() (string, error)
		GroupTopic() (string, error)
	}

	SessionDirectoryTopic string

	SessionDirectorySubscriber interface {
		OnSessionTopic() Subscriber[SessionTopicMessage]
		OnSessionTopicAll() Subscriber[SessionTopicMessage]
		OnGroupTopic() Subscriber[GroupTopicMessage]
	}

	SessionDirectory interface {
		PxGridService

		GetSessions(startTimestamp string, filter any) CallFinalizer[*[]Session]
		GetSessionsForRecovery(startTimestamp, endTimestamp string) CallFinalizer[*[]Session]
		GetSessionByIPAddress(ipAddress string) CallFinalizer[*Session]
		GetSessionByMacAddress(macAddress string) CallFinalizer[*Session]
		GetUserGroups(filter any) CallFinalizer[*[]Group]
		GetUserGroupByUserName(userName string) CallFinalizer[*[]Group]

		SessionDirectorySubscriber

		Properties() SessionDirectoryPropsProvider
	}

	pxGridSessionDirectory struct {
		pxGridService
	}
)

const (
	SessionStateAuthenticating SessionState = "AUTHENTICATING"
	SessionStateAuthenticated  SessionState = "AUTHENTICATED"
	SessionStatePostured       SessionState = "POSTURED"
	SessionStateStarted        SessionState = "STARTED"
	SessionStateDisconnected   SessionState = "DISCONNECTED"

	GroupTypeActiveDirectory            GroupType = "ACTIVE_DIRECTORY"
	GroupTypeIdentity                   GroupType = "IDENTITY"
	GroupTypeExternal                   GroupType = "EXTERNAL"
	GroupTypeInterestingActiveDirectory GroupType = "INTERESTING_ACTIVE_DIRECTORY"

	SessionDirectoryTopicSession    SessionDirectoryTopic = "sessionTopic"
	SessionDirectoryTopicSessionAll SessionDirectoryTopic = "sessionTopicAll"
	SessionDirectoryTopicGroup      SessionDirectoryTopic = "groupTopic"
)

func NewPxGridSessionDirectory(ctrl *PxGridConsumer) SessionDirectory {
	return &pxGridSessionDirectory{
		pxGridService{
			name: "com.cisco.ise.session",
			ctrl: ctrl,
		},
	}
}

// GetSessions retrieves the sessions from the session directory service
func (s *pxGridSessionDirectory) GetSessions(startTimestamp string, filter any) CallFinalizer[*[]Session] {
	payload := map[string]any{}
	if startTimestamp != "" {
		payload["startTimestamp"] = startTimestamp
	}

	if filter != nil {
		payload["filter"] = filter
	}

	type response struct {
		Sessions []Session `json:"sessions"`
	}

	return newCall[*[]Session](
		&s.pxGridService,
		"getSessions",
		payload,
		func(r *Response) (*[]Session, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			return &r.Result.(*response).Sessions, nil
		},
	)
}

// GetSessionsForRecovery retrieves the sessions for recovery from the session directory service
func (s *pxGridSessionDirectory) GetSessionsForRecovery(startTimestamp, endTimestamp string) CallFinalizer[*[]Session] {
	payload := map[string]any{}

	if startTimestamp != "" {
		payload["startTimestamp"] = startTimestamp
	}
	if endTimestamp != "" {
		payload["endTimestamp"] = endTimestamp
	}

	type response struct {
		Sessions []Session `json:"sessions"`
	}

	return newCall[*[]Session](
		&s.pxGridService,
		"getSessionsForRecovery",
		payload,
		func(r *Response) (*[]Session, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			return &r.Result.(*response).Sessions, nil
		},
	)
}

// GetSessionByIPAddress retrieves a session by its IP address
func (s *pxGridSessionDirectory) GetSessionByIPAddress(ipAddress string) CallFinalizer[*Session] {
	if ipAddress == "" {
		return newFailedCall[*Session](ErrInvalidInput)
	}

	return newCall[*Session](
		&s.pxGridService,
		"getSessionByIPAddress",
		map[string]any{"ipAddress": ipAddress},
		simpleResultMapper[*Session],
	)
}

// GetSessionByMacAddress retrieves a session by its MAC address
func (s *pxGridSessionDirectory) GetSessionByMacAddress(macAddress string) CallFinalizer[*Session] {
	if macAddress == "" {
		return newFailedCall[*Session](ErrInvalidInput)
	}

	return newCall[*Session](
		&s.pxGridService,
		"getSessionByMacAddress",
		map[string]any{"macAddress": macAddress},
		simpleResultMapper[*Session],
	)
}

// GetUserGroups retrieves the user groups from the session directory service
func (s *pxGridSessionDirectory) GetUserGroups(filter any) CallFinalizer[*[]Group] {
	payload := map[string]any{}
	if filter != nil {
		payload["filter"] = filter
	}

	type response struct {
		Groups []Group `json:"userGroups"`
	}

	return newCall[*[]Group](
		&s.pxGridService,
		"getUserGroups",
		payload,
		func(r *Response) (*[]Group, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			return &r.Result.(*response).Groups, nil
		},
	)
}

// GetUserGroupByUserName retrieves a user group by its user name
func (s *pxGridSessionDirectory) GetUserGroupByUserName(userName string) CallFinalizer[*[]Group] {
	if userName == "" {
		return newFailedCall[*[]Group](ErrInvalidInput)
	}

	type response struct {
		Groups []Group `json:"groups"`
	}

	return newCall[*[]Group](
		&s.pxGridService,
		"getUserGroupByUserName",
		map[string]any{"userName": userName},
		func(r *Response) (*[]Group, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			if r.StatusCode == 204 {
				return &[]Group{}, nil
			}
			return &r.Result.(*response).Groups, nil
		},
	)
}

func (s *pxGridSessionDirectory) Properties() SessionDirectoryPropsProvider {
	return s
}

func (s *pxGridSessionDirectory) RestBaseURL() (string, error) {
	return s.nodes.GetPropertyString("restBaseUrl")
}

func (s *pxGridSessionDirectory) WSPubsubService() (string, error) {
	return s.nodes.GetPropertyString("wsPubsubService")
}

func (s *pxGridSessionDirectory) SessionTopic() (string, error) {
	return s.nodes.GetPropertyString(string(SessionDirectoryTopicSession))
}

func (s *pxGridSessionDirectory) SessionTopicAll() (string, error) {
	return s.nodes.GetPropertyString(string(SessionDirectoryTopicSessionAll))
}

func (s *pxGridSessionDirectory) GroupTopic() (string, error) {
	return s.nodes.GetPropertyString(string(SessionDirectoryTopicGroup))
}

func (s *pxGridSessionDirectory) OnSessionTopic() Subscriber[SessionTopicMessage] {
	return newSubscriber[SessionTopicMessage](
		&s.pxGridService,
		string(SessionDirectoryTopicSession),
		s.WSPubsubService,
	)
}

func (s *pxGridSessionDirectory) OnSessionTopicAll() Subscriber[SessionTopicMessage] {
	return newSubscriber[SessionTopicMessage](
		&s.pxGridService,
		string(SessionDirectoryTopicSessionAll),
		s.WSPubsubService,
	)
}

func (s *pxGridSessionDirectory) OnGroupTopic() Subscriber[GroupTopicMessage] {
	return newSubscriber[GroupTopicMessage](
		&s.pxGridService,
		string(SessionDirectoryTopicGroup),
		s.WSPubsubService,
	)
}
