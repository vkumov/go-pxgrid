package gopxgrid

import (
	"fmt"
)

type (
	TrustSecConfigurationPropsProvider interface {
		RestBaseURL() (string, error)
		WSPubsubService() (string, error)
		SecurityGroupTopic() (string, error)
		SecurityGroupACLTopic() (string, error)
		SecurityGroupVNVlanTopic() (string, error)
		VirtualNetworkTopic() (string, error)
		EgressPolicyTopic() (string, error)
	}

	EgressMatrix struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		MonitorAll  bool   `json:"monitorAll"`
	}

	EgressPolicy struct {
		ID                         string   `json:"id"`
		Name                       string   `json:"name"`
		MatrixId                   string   `json:"matrixId"`
		Status                     string   `json:"status"`
		Description                string   `json:"description"`
		SourceSecurityGroupID      string   `json:"sourceSecurityGroupId"`
		DestinationSecurityGroupID string   `json:"destinationSecurityGroupId"`
		SGACLIDs                   []string `json:"sgaclIds"`
		Timestamp                  string   `json:"timestamp"`
	}

	VirtualNetwork struct {
		ID                   string `json:"id"`
		Name                 string `json:"name"`
		AdditionalAttributes string `json:"additionalAttributes"`
		Timestamp            string `json:"timestamp"`
	}

	SecurityGroupACL struct {
		ID              string `json:"id"`
		IsDeleted       bool   `json:"isDeleted"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		IPVersion       string `json:"ipVersion"`
		ACL             string `json:"acl"`
		ModelledContent any    `json:"modelledContent"`
		GenerationID    string `json:"generationId"`
		Timestamp       string `json:"timestamp"`
	}

	SecurityGroup struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Tag         int    `json:"tag"`
		Timestamp   string `json:"timestamp"`
	}

	GetSecurityGroupsResponse struct {
		TotalCount            int             `json:"totalCount"`
		SecurityGroups        []SecurityGroup `json:"securityGroups"`
		DeletedSecurityGroups []SecurityGroup `json:"deletedSecurityGroups"`
	}

	GetSecurityGroupACLsResponse struct {
		TotalCount              int                `json:"totalCount"`
		SecurityGroupACLs       []SecurityGroupACL `json:"securityGroupAcls"`
		DeleteSecurityGroupACLs []SecurityGroupACL `json:"deletedSecurityGroupAcls"`
	}

	GetVirtualNetworksResponse struct {
		TotalCount             int              `json:"totalCount"`
		VirtualNetworks        []VirtualNetwork `json:"virtualNetworks"`
		DeletedVirtualNetworks []VirtualNetwork `json:"deletedVirtualNetworks"`
	}

	GetEgressPoliciesResponse struct {
		TotalCount            int            `json:"totalCount"`
		EgressPolicies        []EgressPolicy `json:"egressPolicies"`
		DeletedEgressPolicies []EgressPolicy `json:"deletedEgressPolicies"`
	}

	SecurityGroupTopicMessage struct {
		Sequence      int           `json:"sequence"`
		OperationType OperationType `json:"operation"`
		SecurityGroup SecurityGroup `json:"securityGroup"`
	}

	SecurityGroupACLTopicMessage struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		IPVersion       string `json:"ipVersion"`
		ACL             string `json:"acl"`
		ModelledContent any    `json:"modelledContent"`
		GenerationID    string `json:"generationId"`
		IsReadOnly      bool   `json:"isReadOnly"`
		Sequence        int    `json:"sequence"`
		Deleted         bool   `json:"deleted"`
		Timestamp       string `json:"timestamp"`
	}

	SecurityGroupVNVlanTopicMessage any

	VirtualNetworkTopicMessage struct {
		ID                   string `json:"id"`
		Name                 string `json:"name"`
		AdditionalAttributes string `json:"additionalAttributes"`
		Sequence             int    `json:"sequence"`
		Deleted              bool   `json:"deleted"`
		Timestamp            string `json:"timestamp"`
	}

	EgressPolicyTopicMessage struct {
		ID                 string   `json:"id"`
		Name               string   `json:"name"`
		Description        string   `json:"description"`
		SourceSGTID        string   `json:"sourceSgtId"`
		SourceSGTName      string   `json:"sourceSgtName"`
		DestinationSGTID   string   `json:"destinationSgtId"`
		DestinationSGTName string   `json:"destinationSgtName"`
		MatrixCellStatus   string   `json:"matrixCellStatus"`
		SGACLIDs           []string `json:"sgaclIds"`
		DefaultRule        string   `json:"defaultRule"`
		Sequence           int      `json:"sequence"`
		Deleted            bool     `json:"deleted"`
		Timestamp          string   `json:"timestamp"`
	}

	TrustSecConfigurationTopic string

	TrustSecConfigurationSubscriber interface {
		OnSecurityGroupTopic() Subscriber[SecurityGroupTopicMessage]
		OnSecurityGroupACLTopic() Subscriber[SecurityGroupACLTopicMessage]
		OnSecurityGroupVNVlanTopic() Subscriber[SecurityGroupVNVlanTopicMessage]
		OnVirtualNetworkTopic() Subscriber[VirtualNetworkTopicMessage]
		OnEgressPolicyTopic() Subscriber[EgressPolicyTopicMessage]
	}

	TrustSecConfigurationRest interface {
		GetSecurityGroups(filters ...TrustSecConfigurationRequestFilter) CallFinalizer[*GetSecurityGroupsResponse]
		GetSecurityGroupACLs(filters ...TrustSecConfigurationRequestFilter) CallFinalizer[*GetSecurityGroupACLsResponse]
		GetVirtualNetwork(filters ...TrustSecConfigurationRequestFilter) CallFinalizer[*GetVirtualNetworksResponse]
		GetEgressPolicies(filters ...TrustSecEgressPoliciesRequestFilter) CallFinalizer[*GetEgressPoliciesResponse]
		GetEgressMatrices() CallFinalizer[*[]EgressMatrix]
	}

	TrustSecConfiguration interface {
		PxGridService

		Rest() TrustSecConfigurationRest

		TrustSecConfigurationSubscriber

		Properties() TrustSecConfigurationPropsProvider
	}

	pxGridTrustSecConfiguration struct {
		pxGridService
	}
)

const (
	TrustSecConfigurationTopicSecurityGroup       TrustSecConfigurationTopic = "securityGroupTopic"
	TrustSecConfigurationTopicSecurityGroupACL    TrustSecConfigurationTopic = "securityGroupAclTopic"
	TrustSecConfigurationTopicSecurityGroupVNVlan TrustSecConfigurationTopic = "securityGroupVnVlanTopic"
	TrustSecConfigurationTopicVirtualNetwork      TrustSecConfigurationTopic = "virtualnetworkTopic"
	TrustSecConfigurationTopicEgressPolicy        TrustSecConfigurationTopic = "egressPolicyTopic"

	TrustSecConfigurationServiceName = "com.cisco.ise.config.trustsec"
)

func NewPxGridTrustSecConfiguration(ctrl *PxGridConsumer) TrustSecConfiguration {
	return &pxGridTrustSecConfiguration{
		pxGridService{
			name: TrustSecConfigurationServiceName,
			ctrl: ctrl,
			log:  ctrl.cfg.Logger.With("svc", TrustSecConfigurationServiceName),
		},
	}
}

func (t *pxGridTrustSecConfiguration) Rest() TrustSecConfigurationRest {
	return t
}

func (t *pxGridTrustSecConfiguration) Properties() TrustSecConfigurationPropsProvider {
	return t
}

func (t *pxGridTrustSecConfiguration) RestBaseURL() (string, error) {
	return t.nodes.GetPropertyString("restBaseUrl")
}

func (t *pxGridTrustSecConfiguration) WSPubsubService() (string, error) {
	return t.nodes.GetPropertyString("wsPubsubService")
}

func (t *pxGridTrustSecConfiguration) SecurityGroupTopic() (string, error) {
	return t.nodes.GetPropertyString(string(TrustSecConfigurationTopicSecurityGroup))
}

func (t *pxGridTrustSecConfiguration) SecurityGroupACLTopic() (string, error) {
	return t.nodes.GetPropertyString(string(TrustSecConfigurationTopicSecurityGroupACL))
}

func (t *pxGridTrustSecConfiguration) SecurityGroupVNVlanTopic() (string, error) {
	return t.nodes.GetPropertyString(string(TrustSecConfigurationTopicSecurityGroupVNVlan))
}

func (t *pxGridTrustSecConfiguration) VirtualNetworkTopic() (string, error) {
	return t.nodes.GetPropertyString(string(TrustSecConfigurationTopicVirtualNetwork))
}

func (t *pxGridTrustSecConfiguration) EgressPolicyTopic() (string, error) {
	return t.nodes.GetPropertyString(string(TrustSecConfigurationTopicEgressPolicy))
}

type (
	trustSecConfigurationRequestFilter struct {
		ID             string `json:"id,omitempty"`
		StartIndex     *int   `json:"startIndex,omitempty"`
		RecordCount    *int   `json:"recordCount,omitempty"`
		StartTimestamp string `json:"startTimestamp,omitempty"`
		EndTimestamp   string `json:"endTimestamp,omitempty"`
	}

	TrustSecConfigurationRequestFilter func(*trustSecConfigurationRequestFilter)
)

func WithID(id string) TrustSecConfigurationRequestFilter {
	return func(f *trustSecConfigurationRequestFilter) {
		f.ID = id
	}
}

func WithStartIndex(startIndex int) TrustSecConfigurationRequestFilter {
	return func(f *trustSecConfigurationRequestFilter) {
		f.StartIndex = &startIndex
	}
}

func WithRecordCount(recordCount int) TrustSecConfigurationRequestFilter {
	return func(f *trustSecConfigurationRequestFilter) {
		f.RecordCount = &recordCount
	}
}

func WithStartTimestamp(startTimestamp string) TrustSecConfigurationRequestFilter {
	return func(f *trustSecConfigurationRequestFilter) {
		f.StartTimestamp = startTimestamp
	}
}

func WithEndTimestamp(endTimestamp string) TrustSecConfigurationRequestFilter {
	return func(f *trustSecConfigurationRequestFilter) {
		f.EndTimestamp = endTimestamp
	}
}

func applyTrustSecConfigFilters(f *trustSecConfigurationRequestFilter, filters []TrustSecConfigurationRequestFilter) {
	for _, filter := range filters {
		filter(f)
	}
}

func (t *pxGridTrustSecConfiguration) GetSecurityGroups(filters ...TrustSecConfigurationRequestFilter) CallFinalizer[*GetSecurityGroupsResponse] {
	var payload any
	if len(filters) > 0 {
		f := &trustSecConfigurationRequestFilter{}
		applyTrustSecConfigFilters(f, filters)
		payload = f
	} else {
		payload = map[string]any{}
	}

	return newCall[*GetSecurityGroupsResponse](
		&t.pxGridService,
		"getSecurityGroups",
		payload,
		simpleResultMapper[*GetSecurityGroupsResponse],
	)
}

func (t *pxGridTrustSecConfiguration) GetSecurityGroupACLs(filters ...TrustSecConfigurationRequestFilter) CallFinalizer[*GetSecurityGroupACLsResponse] {
	var payload any
	if len(filters) > 0 {
		f := &trustSecConfigurationRequestFilter{}
		applyTrustSecConfigFilters(f, filters)
		payload = f
	} else {
		payload = map[string]any{}
	}

	return newCall[*GetSecurityGroupACLsResponse](
		&t.pxGridService,
		"getSecurityGroupAcls",
		payload,
		simpleResultMapper[*GetSecurityGroupACLsResponse],
	)
}

func (t *pxGridTrustSecConfiguration) GetVirtualNetwork(filters ...TrustSecConfigurationRequestFilter) CallFinalizer[*GetVirtualNetworksResponse] {
	var payload any
	if len(filters) > 0 {
		f := &trustSecConfigurationRequestFilter{}
		applyTrustSecConfigFilters(f, filters)
		payload = f
	} else {
		payload = map[string]any{}
	}

	return newCall[*GetVirtualNetworksResponse](
		&t.pxGridService,
		"getVirtualNetwork",
		payload,
		simpleResultMapper[*GetVirtualNetworksResponse],
	)
}

type (
	trustSecEgressPoliciesRequestFilter struct {
		trustSecConfigurationRequestFilter
		MatrixID string `json:"matrixId,omitempty"`
	}

	TrustSecEgressPoliciesRequestFilter func(*trustSecEgressPoliciesRequestFilter)
)

func WithEgressPolicyID(id string) TrustSecEgressPoliciesRequestFilter {
	return func(f *trustSecEgressPoliciesRequestFilter) {
		f.ID = id
	}
}

func WithEgressPolicyStartIndex(startIndex int) TrustSecEgressPoliciesRequestFilter {
	return func(f *trustSecEgressPoliciesRequestFilter) {
		f.StartIndex = &startIndex
	}
}

func WithEgressPolicyRecordCount(recordCount int) TrustSecEgressPoliciesRequestFilter {
	return func(f *trustSecEgressPoliciesRequestFilter) {
		f.RecordCount = &recordCount
	}
}

func WithEgressPolicyStartTimestamp(startTimestamp string) TrustSecEgressPoliciesRequestFilter {
	return func(f *trustSecEgressPoliciesRequestFilter) {
		f.StartTimestamp = startTimestamp
	}
}

func WithEgressPolicyEndTimestamp(endTimestamp string) TrustSecEgressPoliciesRequestFilter {
	return func(f *trustSecEgressPoliciesRequestFilter) {
		f.EndTimestamp = endTimestamp
	}
}

func WithEgressPolicyMatrixID(matrixID string) TrustSecEgressPoliciesRequestFilter {
	return func(f *trustSecEgressPoliciesRequestFilter) {
		f.MatrixID = matrixID
	}
}

func applyTrustSecEgressPoliciesFilters(f *trustSecEgressPoliciesRequestFilter, filters []TrustSecEgressPoliciesRequestFilter) {
	for _, filter := range filters {
		filter(f)
	}
}

func (t *pxGridTrustSecConfiguration) GetEgressPolicies(filters ...TrustSecEgressPoliciesRequestFilter) CallFinalizer[*GetEgressPoliciesResponse] {
	var payload any
	if len(filters) > 0 {
		f := &trustSecEgressPoliciesRequestFilter{}
		applyTrustSecEgressPoliciesFilters(f, filters)
		payload = f
	} else {
		payload = map[string]any{}
	}

	return newCall[*GetEgressPoliciesResponse](
		&t.pxGridService,
		"getEgressPolicies",
		payload,
		simpleResultMapper[*GetEgressPoliciesResponse],
	)
}

func (t *pxGridTrustSecConfiguration) GetEgressMatrices() CallFinalizer[*[]EgressMatrix] {
	type response struct {
		EgressMatrices []EgressMatrix `json:"egressMatrices"`
	}

	return newCall[*[]EgressMatrix](
		&t.pxGridService,
		"getEgressMatrices",
		map[string]any{},
		func(r *Response) (*[]EgressMatrix, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			if r.StatusCode == 204 {
				return &[]EgressMatrix{}, nil
			}
			return &r.Result.(*response).EgressMatrices, nil
		},
	)
}

func (t *pxGridTrustSecConfiguration) OnSecurityGroupTopic() Subscriber[SecurityGroupTopicMessage] {
	return newSubscriber[SecurityGroupTopicMessage](
		&t.pxGridService,
		string(TrustSecConfigurationTopicSecurityGroup),
		t.WSPubsubService,
	)
}

func (t *pxGridTrustSecConfiguration) OnSecurityGroupACLTopic() Subscriber[SecurityGroupACLTopicMessage] {
	return newSubscriber[SecurityGroupACLTopicMessage](
		&t.pxGridService,
		string(TrustSecConfigurationTopicSecurityGroupACL),
		t.WSPubsubService,
	)
}

func (t *pxGridTrustSecConfiguration) OnSecurityGroupVNVlanTopic() Subscriber[SecurityGroupVNVlanTopicMessage] {
	return newSubscriber[SecurityGroupVNVlanTopicMessage](
		&t.pxGridService,
		string(TrustSecConfigurationTopicSecurityGroupVNVlan),
		t.WSPubsubService,
	)
}

func (t *pxGridTrustSecConfiguration) OnVirtualNetworkTopic() Subscriber[VirtualNetworkTopicMessage] {
	return newSubscriber[VirtualNetworkTopicMessage](
		&t.pxGridService,
		string(TrustSecConfigurationTopicVirtualNetwork),
		t.WSPubsubService,
	)
}

func (t *pxGridTrustSecConfiguration) OnEgressPolicyTopic() Subscriber[EgressPolicyTopicMessage] {
	return newSubscriber[EgressPolicyTopicMessage](
		&t.pxGridService,
		string(TrustSecConfigurationTopicEgressPolicy),
		t.WSPubsubService,
	)
}
