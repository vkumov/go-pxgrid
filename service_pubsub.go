package gopxgrid

type (
	PubSubPropsProvider interface {
		WSURL() (string, error)
	}

	PubSub interface {
		PxGridService

		Properties() PubSubPropsProvider
	}

	pxGridPubSub struct {
		pxGridService
	}
)

func NewPxGridPubSub(ctrl Controller) PubSub {
	return &pxGridPubSub{
		pxGridService{
			name: "com.cisco.ise.pubsub",
			ctrl: ctrl,
		},
	}
}

func (p *pxGridPubSub) Properties() PubSubPropsProvider {
	return p
}

func (p *pxGridPubSub) WSURL() (string, error) {
	return p.nodes.GetPropertyString("wsUrl")
}
