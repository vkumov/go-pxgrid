package gopxgrid

import "errors"

type (
	ANCKeyValue struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	ANCAsset struct {
		AssetId               string        `json:"assetId"`
		AssetName             string        `json:"assetName"`
		AssetIPAddress        string        `json:"assetIpAddress"`
		AssetMACAddress       string        `json:"assetMacAddress"`
		AssetVendor           string        `json:"assetVendor"`
		AssetProductID        string        `json:"assetProductId"`
		AssetSerialNumber     string        `json:"assetSerialNumber"`
		AssetDeviceType       string        `json:"assetDeviceType"`
		AssetSWRevision       string        `json:"assetSwRevision"`
		AssetHWRevision       string        `json:"assetHwRevision"`
		AssetProtocol         string        `json:"assetProtocol"`
		AssetCustomAttributes []ANCKeyValue `json:"assetCustomAttributes"`
		AssetConnectedLinks   []ANCKeyValue `json:"assetConnectedLinks"`
	}

	ANCAssetTopicMessage struct {
		OperationType OperationType `json:"opType"`
		Asset         ANCAsset      `json:"asset"`
	}

	EndpointAssetPropsProvider interface {
		WSPubsubService() (string, error)
		AssetTopic() (string, error)
	}

	EndpointAssetSubscriber interface {
		OnAssetTopic() (*Subscription[ANCAssetTopicMessage], error)
	}

	EndpointAsset interface {
		PxGridService

		Subscribe() EndpointAssetSubscriber

		Properties() EndpointAssetPropsProvider
	}

	pxGridEndpointAsset struct {
		pxGridService
	}
)

func NewPxGridEndpointAsset(ctrl Controller) EndpointAsset {
	return &pxGridEndpointAsset{
		pxGridService{
			name: "com.cisco.endpoint.asset",
			ctrl: ctrl,
		},
	}
}

func (e *pxGridEndpointAsset) Properties() EndpointAssetPropsProvider {
	return e
}

func (e *pxGridEndpointAsset) WSPubsubService() (string, error) {
	return e.nodes.GetPropertyString("wsPubsubService")
}

func (e *pxGridEndpointAsset) AssetTopic() (string, error) {
	return e.nodes.GetPropertyString("assetTopic")
}

func (e *pxGridEndpointAsset) Subscribe() EndpointAssetSubscriber {
	return e
}

func (e *pxGridEndpointAsset) OnAssetTopic() (*Subscription[ANCAssetTopicMessage], error) {
	return nil, errors.New("not implemented")
}
