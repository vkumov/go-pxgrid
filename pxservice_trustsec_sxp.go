package gopxgrid

import (
	"fmt"
)

type (
	TrustSecSXPBinding struct {
		Tag          string `json:"tag"`
		IPPrefix     string `json:"ipPrefix"`
		Source       string `json:"source"`
		PeerSequence string `json:"peerSequence"`
		VPN          string `json:"vpn"`
	}

	TrustSecSXPBindingTopicMessage struct {
		OperationType OperationType      `json:"operation"`
		Binding       TrustSecSXPBinding `json:"binding"`
	}

	TrustSecSXPPropsProvider interface {
		RestBaseURL() (string, error)
		WSPubsubService() (string, error)
		BindingTopic() (string, error)
	}

	TrustSecSXPTopic string

	TrustSecSXPSubscriber interface {
		OnBindingTopic() Subscriber[TrustSecSXPBindingTopicMessage]
	}

	TrustSecSXPRest interface {
		GetBindings(filter any) CallFinalizer[*[]TrustSecSXPBinding]
	}

	TrustSecSXP interface {
		PxGridService

		Rest() TrustSecSXPRest

		TrustSecSXPSubscriber

		Properties() TrustSecSXPPropsProvider
	}

	pxGridTrustSecSXP struct {
		pxGridService
	}
)

const (
	TrustSecSXPTopicBinding TrustSecSXPTopic = "bindingTopic"

	TrustSecSXPServiceName = "com.cisco.ise.sxp"
)

func NewPxGridTrustSecSXP(ctrl *PxGridConsumer) TrustSecSXP {
	return &pxGridTrustSecSXP{
		pxGridService{
			name: TrustSecSXPServiceName,
			ctrl: ctrl,
			log:  ctrl.cfg.Logger.With("svc", TrustSecSXPServiceName),
		},
	}
}

func (t *pxGridTrustSecSXP) Rest() TrustSecSXPRest {
	return t
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
	return t.nodes.GetPropertyString(string(TrustSecSXPTopicBinding))
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

func (t *pxGridTrustSecSXP) OnBindingTopic() Subscriber[TrustSecSXPBindingTopicMessage] {
	return newSubscriber[TrustSecSXPBindingTopicMessage](
		&t.pxGridService,
		string(TrustSecSXPTopicBinding),
		t.WSPubsubService,
	)
}
