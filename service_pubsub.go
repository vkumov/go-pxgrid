package gopxgrid

import (
	"context"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-stomp/stomp/v3"
	"github.com/gorilla/websocket"
)

type (
	PubSubPropsProvider interface {
		WSURL() (string, error)
	}

	PubSubSubscriber interface {
		Subscribe(ctx context.Context, topic string, node ServiceNode) (*stomp.Subscription, error)
	}

	PubSub interface {
		PxGridService

		PubSubSubscriber

		Properties() PubSubPropsProvider
	}

	PubSubEndpoint struct {
		dialer websocket.Dialer
		ws     *websocket.Conn
		stomp  *stomp.Conn
		ticker *time.Ticker
		wsURL  string

		nodeName string
		secret   string
	}

	pxGridPubSub struct {
		pxGridService

		eps     map[string]*PubSubEndpoint
		epMutex sync.RWMutex
	}
)

func NewPxGridPubSub(ctrl *PxGridConsumer) PubSub {
	return &pxGridPubSub{
		pxGridService: pxGridService{
			name: "com.cisco.ise.pubsub",
			ctrl: ctrl,
		},
		eps: make(map[string]*PubSubEndpoint),
	}
}

func (p *pxGridPubSub) Properties() PubSubPropsProvider {
	return p
}

func (p *pxGridPubSub) WSURL() (string, error) {
	return p.nodes.GetPropertyString("wsUrl")
}

func (p *pxGridPubSub) Subscribe(ctx context.Context, topic string, node ServiceNode) (*stomp.Subscription, error) {
	ep, err := p.getEndpoint(node)
	if err != nil {
		return nil, err
	}

	err = ep.connect(ctx)
	if err != nil {
		return nil, err
	}

	return ep.stomp.Subscribe(topic, stomp.AckAuto)
}

func (p *pxGridPubSub) createEndpoint(wsURL, nodeName, secret string) *PubSubEndpoint {
	ep := &PubSubEndpoint{
		dialer: websocket.Dialer{
			TLSClientConfig: p.ctrl.ClientTLSConfig(),
			Proxy:           http.ProxyFromEnvironment,
			NetDialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return p.ctrl.DialContext(ctx, network, addr)
			},
		},
		wsURL:    wsURL,
		nodeName: nodeName,
		secret:   secret,
	}

	return ep
}

func (p *pxGridPubSub) getEndpoint(node ServiceNode) (*PubSubEndpoint, error) {
	p.epMutex.Lock()
	defer p.epMutex.Unlock()

	var wsURL string
	if raw, ok := node.Properties["wsUrl"]; !ok {
		return nil, ErrPropertyNotFound
	} else if wsURL, ok = raw.(string); !ok {
		return nil, ErrPropertyNotString
	}

	if wsURL == "" {
		return nil, ErrPropertyNotFound
	}

	ep, ok := p.eps[wsURL]
	if !ok {
		ep = p.createEndpoint(wsURL, node.NodeName, node.Secret)
		p.eps[wsURL] = ep
	}

	return ep, nil
}

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

func (e *PubSubEndpoint) pinger(ch <-chan time.Time) {
	for range ch {
		err := e.ws.WriteControl(websocket.PingMessage, []byte(""), time.Time{})
		if err != nil {
			return
		}
	}
}

func (e *PubSubEndpoint) getAuthHeaders() http.Header {
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte(e.nodeName+":"+e.secret))
	return http.Header{"Authorization": {basic}}
}

func (e *PubSubEndpoint) connect(ctx context.Context) (err error) {
	e.ws, _, err = e.dialer.DialContext(ctx, e.wsURL, e.getAuthHeaders())
	if err != nil {
		return err
	}

	e.stomp, err = stomp.Connect(e.ws.NetConn())
	if err != nil {
		return errors.Join(err, e.ws.Close())
	}

	e.ws.SetPongHandler(func(string) error {
		e.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	e.ticker = time.NewTicker(pingPeriod)
	go e.pinger(e.ticker.C)

	return
}

func (e *PubSubEndpoint) Disconnect() error {
	e.ticker.Stop()
	return e.ws.Close()
}
