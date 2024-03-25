package gopxgrid

import "fmt"

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

	TrustSecOperation string

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

	TrustSecConfiguration interface {
		PxGridService

		GetSecurityGroups(filters ...TrustSecConfigurationRequestFilter) CallFinalizer[*GetSecurityGroupsResponse]
		GetSecurityGroupACLs(filters ...TrustSecConfigurationRequestFilter) CallFinalizer[*GetSecurityGroupACLsResponse]
		GetVirtualNetwork(filters ...TrustSecConfigurationRequestFilter) CallFinalizer[*GetVirtualNetworksResponse]
		GetEgressPolicies(filters ...TrustSecEgressPoliciesRequestFilter) CallFinalizer[*GetEgressPoliciesResponse]
		GetEgressMatrices() CallFinalizer[*[]EgressMatrix]

		Properties() TrustSecConfigurationPropsProvider
	}

	pxGridTrustSecConfiguration struct {
		pxGridService
	}
)

const (
	TrustSecOperationCreate TrustSecOperation = "CREATE"
	TrustSecOperationUpdate TrustSecOperation = "UPDATE"
	TrustSecOperationDelete TrustSecOperation = "DELETE"
)

func NewPxGridTrustSecConfiguration(ctrl Controller) TrustSecConfiguration {
	return &pxGridTrustSecConfiguration{
		pxGridService{
			name: "com.cisco.ise.config.trustsec",
			ctrl: ctrl,
		},
	}
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
	return t.nodes.GetPropertyString("securityGroupTopic")
}

func (t *pxGridTrustSecConfiguration) SecurityGroupACLTopic() (string, error) {
	return t.nodes.GetPropertyString("securityGroupAclTopic")
}

func (t *pxGridTrustSecConfiguration) SecurityGroupVNVlanTopic() (string, error) {
	return t.nodes.GetPropertyString("securityGroupVnVlanTopic")
}

func (t *pxGridTrustSecConfiguration) VirtualNetworkTopic() (string, error) {
	return t.nodes.GetPropertyString("virtualnetworkTopic")
}

func (t *pxGridTrustSecConfiguration) EgressPolicyTopic() (string, error) {
	return t.nodes.GetPropertyString("egressPolicyTopic")
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
