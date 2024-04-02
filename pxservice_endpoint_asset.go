package gopxgrid

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

	EndpointAssetTopic string

	EndpointAssetSubscriber interface {
		OnAssetTopic() Subscriber[ANCAssetTopicMessage]
	}

	EndpointAsset interface {
		PxGridService

		EndpointAssetSubscriber

		Properties() EndpointAssetPropsProvider
	}

	pxGridEndpointAsset struct {
		pxGridService
	}
)

const (
	EndpointAssetTopicAsset EndpointAssetTopic = "assetTopic"
)

func NewPxGridEndpointAsset(ctrl *PxGridConsumer) EndpointAsset {
	return &pxGridEndpointAsset{
		pxGridService{
			name: "com.cisco.endpoint.asset",
			ctrl: ctrl,
			log:  ctrl.cfg.Logger.With("svc", "com.cisco.endpoint.asset"),
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
	return e.nodes.GetPropertyString(string(EndpointAssetTopicAsset))
}

func (e *pxGridEndpointAsset) OnAssetTopic() Subscriber[ANCAssetTopicMessage] {
	return newSubscriber[ANCAssetTopicMessage](
		&e.pxGridService,
		string(EndpointAssetTopicAsset),
		e.WSPubsubService,
	)
}
