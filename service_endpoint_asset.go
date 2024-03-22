package gopxgrid

type (
	EndpointAsset interface {
		PxGridService
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
