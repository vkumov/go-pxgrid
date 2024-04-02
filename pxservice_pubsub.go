package gopxgrid

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
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
		Subscribe(ctx context.Context, picker ServiceNodePickerFactory, topic string) (*stomp.Subscription, error)
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

		readerBuffer []byte
		writeBuffer  []byte
		log          *slog.Logger
	}

	pxGridPubSub struct {
		pxGridService

		eps     map[string]*PubSubEndpoint
		epMutex sync.RWMutex
	}
)

func NewPxGridPubSub(ctrl *PxGridConsumer, svc string) PubSub {
	ctrl.cfg.Logger.Debug("New pxGrid PubSub service", "service", svc)
	return &pxGridPubSub{
		pxGridService: pxGridService{
			name: svc,
			ctrl: ctrl,
			log:  ctrl.cfg.Logger.WithGroup("pubsub"),
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

func (p *pxGridPubSub) Subscribe(ctx context.Context, picker ServiceNodePickerFactory, topic string) (*stomp.Subscription, error) {
	n := p.orDefaultFactory(picker)(p.nodes)
	for {
		node, more, err := n.PickNode()
		if err != nil {
			return nil, err
		}
		p.log.Debug("PubSub Subscribe", "node", node.NodeName, "secret", node.Secret, "topic", topic)

		ep, err := p.getEndpoint(node)
		if err != nil {
			if !more {
				return nil, err
			}
			continue
		}
		p.log.Debug("WS Endpoint", "wsURL", ep.wsURL)

		err = ep.connect(ctx)
		if err != nil {
			if !more {
				return nil, err
			}
			continue
		}

		return ep.stomp.Subscribe(topic, stomp.AckAuto)
	}
}

func (p *pxGridPubSub) createEndpoint(wsURL, secret string) *PubSubEndpoint {
	p.log.Debug("Create WS PubSub Endpoint", "wsURL", wsURL, "nodeName", p.ctrl.cfg.NodeName)
	ep := &PubSubEndpoint{
		dialer: websocket.Dialer{
			TLSClientConfig: p.ctrl.ClientTLSConfig(),
			Proxy:           http.ProxyFromEnvironment,
			NetDialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return p.ctrl.DialContext(ctx, network, addr)
			},
		},
		wsURL:    wsURL,
		nodeName: p.ctrl.cfg.NodeName,
		secret:   secret,
		log:      p.log.WithGroup("pxgrid.pubsub.endpoint"),
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
		ep = p.createEndpoint(wsURL, node.Secret)
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
		e.log.Debug("Sending ping")
		err := e.ws.WriteControl(websocket.PingMessage, []byte(""), time.Time{})
		if err != nil {
			e.log.Error("Ping failed", "error", err)
			return
		}
	}
}

func (e *PubSubEndpoint) getAuthHeaders() http.Header {
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte(e.nodeName+":"+e.secret))
	return http.Header{"Authorization": {basic}}
}

func (e *PubSubEndpoint) connect(ctx context.Context) (err error) {
	e.log.Debug("WebSocket dial", "wsURL", e.wsURL)
	e.ws, _, err = e.dialer.DialContext(ctx, e.wsURL, e.getAuthHeaders())
	if err != nil {
		return err
	}

	e.log.Debug("STOMP connect")
	e.stomp, err = stomp.Connect(e, stomp.ConnOpt.HeartBeat(0, 0))
	if err != nil {
		return errors.Join(err, e.ws.Close())
	}

	e.log.Debug("STOMP connected, setting up ping/pong")
	e.ws.SetPongHandler(func(string) error {
		e.log.Debug("Received pong")
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

func (e *PubSubEndpoint) Read(p []byte) (int, error) {
	// if we have no more data, read the next message from the websocket
	if len(e.readerBuffer) == 0 {
		_, msg, err := e.ws.ReadMessage()
		if err != nil {
			return 0, err
		}
		e.readerBuffer = msg
	}

	n := copy(p, e.readerBuffer)
	e.readerBuffer = e.readerBuffer[n:]
	return n, nil
}

func (e *PubSubEndpoint) Write(p []byte) (int, error) {
	var err error
	e.writeBuffer = append(e.writeBuffer, p...)
	// if we reach a null byte or the entire message is a newline (heartbeat), send the message
	if p[len(p)-1] == 0x00 || (len(e.writeBuffer) == 1 && len(p) == 1 && p[0] == 0x0a) {
		err = e.ws.WriteMessage(websocket.BinaryMessage, e.writeBuffer)
		e.writeBuffer = []byte{}
	}
	return len(p), err
}

func (e *PubSubEndpoint) Close() error {
	return e.ws.Close()
}
