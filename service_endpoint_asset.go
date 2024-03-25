package gopxgrid

type (
	EndpointAssetPropsProvider interface {
		WSPubsubService() (string, error)
		AssetTopic() (string, error)
	}

	EndpointAsset interface {
		PxGridService

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
