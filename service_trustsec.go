package gopxgrid

import "errors"

type (
	Policy struct {
		SourceSGT                  int    `json:"sourceSgt"`
		SourceSGTGenerationID      string `json:"sourceSgtGenerationId"`
		DestinationSGT             int    `json:"destinationSgt"`
		DestinationSGTGenerationID string `json:"destinationSgtGenerationId"`
		SGACLName                  string `json:"sgaclName"`
		SGACLGenerationID          string `json:"sgaclGenerationId"`
	}

	PolicyDownloadStatus string

	PolicyDownload struct {
		Timestamp       string               `json:"timestamp"`
		ServerName      string               `json:"serverName"`
		Status          PolicyDownloadStatus `json:"status"`
		FailureReason   string               `json:"failureReason"`
		NASIPAddress    string               `json:"nasIpAddress"`
		MatrixName      string               `json:"matrixName"`
		RBACLSourceList string               `json:"rbaclSourceList"`
		Policies        []Policy             `json:"policies"`
	}

	PolicyDownloadTopicMessage struct {
		PolicyDownloads []PolicyDownload `json:"policyDownloads"`
	}

	TrustSecPropsProvider interface {
		WSPubsubService() (string, error)
		PolicyDownloadTopic() (string, error)
	}

	TrustSecSubscriber interface {
		OnPolicyDownloadTopic() (*Subscription[PolicyDownloadTopicMessage], error)
	}

	TrustSec interface {
		PxGridService

		Subscribe() TrustSecSubscriber

		Properties() TrustSecPropsProvider
	}

	pxGridTrustSec struct {
		pxGridService
	}
)

const (
	PolicyDownloadStatusSuccess PolicyDownloadStatus = "SUCCESS"
	PolicyDownloadStatusFailure PolicyDownloadStatus = "FAILURE"
)

func NewPxGridTrustSec(ctrl Controller) TrustSec {
	return &pxGridTrustSec{
		pxGridService{
			name: "com.cisco.ise.trustsec",
			ctrl: ctrl,
		},
	}
}

func (t *pxGridTrustSec) Properties() TrustSecPropsProvider {
	return t
}

func (t *pxGridTrustSec) WSPubsubService() (string, error) {
	return t.nodes.GetPropertyString("wsPubsubService")
}

func (t *pxGridTrustSec) PolicyDownloadTopic() (string, error) {
	return t.nodes.GetPropertyString("policyDownloadTopic")
}

func (t *pxGridTrustSec) Subscribe() TrustSecSubscriber {
	return t
}

func (t *pxGridTrustSec) OnPolicyDownloadTopic() (*Subscription[PolicyDownloadTopicMessage], error) {
	return nil, errors.New("not implemented")
}
