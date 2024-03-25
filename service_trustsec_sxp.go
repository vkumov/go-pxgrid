package gopxgrid

import "fmt"

type (
	TrustSecSXPBinding struct {
		Tag          string `json:"tag"`
		IPPrefix     string `json:"ipPrefix"`
		Source       string `json:"source"`
		PeerSequence string `json:"peerSequence"`
		VPN          string `json:"vpn"`
	}

	TrustSecSXPPropsProvider interface {
		RestBaseURL() (string, error)
		WSPubsubService() (string, error)
		BindingTopic() (string, error)
	}

	TrustSecSXP interface {
		PxGridService

		GetBindings(filter any) CallFinalizer[*[]TrustSecSXPBinding]

		Properties() TrustSecSXPPropsProvider
	}

	pxGridTrustSecSXP struct {
		pxGridService
	}
)

func NewPxGridTrustSecSXP(ctrl Controller) TrustSecSXP {
	return &pxGridTrustSecSXP{
		pxGridService{
			name: "com.cisco.ise.sxp",
			ctrl: ctrl,
		},
	}
}

func (t *pxGridTrustSecSXP) Properties() TrustSecSXPPropsProvider {
	return t
}

func (t *pxGridTrustSecSXP) RestBaseURL() (string, error) {
	return t.nodes.GetPropertyString("restBaseUrl")
}

func (t *pxGridTrustSecSXP) WSPubsubService() (string, error) {
	return t.nodes.GetPropertyString("wsPubsubService")
}

func (t *pxGridTrustSecSXP) BindingTopic() (string, error) {
	return t.nodes.GetPropertyString("bindingTopic")
}

func (t *pxGridTrustSecSXP) GetBindings(filter any) CallFinalizer[*[]TrustSecSXPBinding] {
	payload := map[string]any{}
	if filter != nil {
		payload["filter"] = filter
	}

	type response struct {
		Bindings []TrustSecSXPBinding `json:"bindings"`
	}

	return newCall[*[]TrustSecSXPBinding](
		&t.pxGridService,
		"getBindings",
		payload,
		func(r *Response) (*[]TrustSecSXPBinding, error) {
			if r.StatusCode > 299 {
				return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
			}
			if r.StatusCode == 204 {
				return &[]TrustSecSXPBinding{}, nil
			}
			return &r.Result.(*response).Bindings, nil
		},
	)
}