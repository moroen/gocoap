package gocoap

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	piondtls "github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v2/dtls"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/udp/client"
	"github.com/plgd-dev/go-coap/v2/udp/message/pool"
	log "github.com/sirupsen/logrus"
)

var _retryLimit uint = 3
var _retryDelay = 1

var _cancel func()
var _ctx context.Context

type CoapDTLSConnection struct {
	mu                 sync.Mutex
	Host               string
	Port               int
	Ident              string
	Key                string
	UseQueue           bool
	OnConnect          func()
	OnDisconnect       func()
	OnCanceled         func()
	OnConnectionFailed func()
	_connection        *client.ClientConn
	_status            int
	queue              []CoapDTLSRequest
}

type CoapDTLSRequest struct {
	RequestMethod string
	Uri           string
	Payload       string
	Handler       func([]byte, error)
	Context       context.Context
	WaitGroup     *sync.WaitGroup
	KeepAlive     int
}

func (c *CoapDTLSConnection) Connect() error {
	if c._status > 0 {
		return nil
	}

	c._status = 1 // Connecting

	_ctx, _cancel = context.WithCancel(context.Background())

	ticker := time.NewTicker(time.Duration(5) * time.Second)
	for {
		if conn, err := dtls.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), &piondtls.Config{
			PSK: func(hint []byte) ([]byte, error) {
				// fmt.Printf("Server's hint: %s \n", hint)
				return []byte(c.Key), nil
			},
			PSKIdentityHint: []byte(c.Ident),
			CipherSuites:    []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
		}); err == nil {
			c._connection = conn
			if c.OnConnect != nil {
				c._status = 2
				c.OnConnect()
			}

			if c.UseQueue {
				c.HandleQueue()
			}

			return nil
		} else {
			if c.OnConnectionFailed != nil {
				c._status = 1
				c.OnConnectionFailed()
			}
		}
		select {
		case <-ticker.C:
			break
		case <-_ctx.Done():
			if c.OnCanceled != nil {
				c.OnCanceled()
			}
			return ConnectionContextCanceled
		}
	}
}

func (c *CoapDTLSConnection) Disconnect() error {
	_cancel()
	if c._status == 2 {
		c._connection.Close()
	}
	c._status = 0
	if c.OnDisconnect != nil {
		c.OnDisconnect()
	}
	return nil
}

func (c *CoapDTLSConnection) HandleError(request CoapDTLSRequest) {
	if c.UseQueue {
		log.WithFields(log.Fields{
			"Uri": request.Uri,
		}).Debug("Adding request to queue")
	}

	c.AddToQueue(request)

	if c._status == 2 {
		c.Disconnect()
	}
	c.Connect()
}

func (c *CoapDTLSConnection) GET(ctx context.Context, uri string, handler func([]byte, error)) {
	log.WithFields(log.Fields{
		"Uri": uri,
	}).Debug("CoapDTLSConnection.GET")

	if c._status != 2 {
		log.WithFields(log.Fields{
			"Error": "Not connected",
		}).Error("COAP - GET")
		c.HandleError(CoapDTLSRequest{RequestMethod: "GET", Uri: uri, Handler: handler})
		return
	}

	if response, err := c._connection.Get(ctx, uri); err == nil {
		if m, err := response.ReadBody(); err == nil {
			handler(m, ProcessMessageCode(response))
		} else {
			handler([]byte{}, err)
		}
	} else {
		log.WithFields(log.Fields{
			"Error": err.Error(),
		}).Error("Coap - GET")
		c.HandleError(CoapDTLSRequest{RequestMethod: "GET", Uri: uri, Handler: handler})

	}
}

func (c *CoapDTLSConnection) PUT(ctx context.Context, uri string, payload string, handler func([]byte, error)) {
	if response, err := c._connection.Put(ctx, uri, message.AppJSON, bytes.NewReader([]byte(payload))); err == nil {
		if m, err := response.ReadBody(); err == nil {
			handler(m, ProcessMessageCode(response))
		} else {
			handler([]byte{}, err)
		}
	} else {
		log.WithFields(log.Fields{
			"Error": err.Error(),
		}).Error("Coap - PUT")
	}
}

func (c *CoapDTLSConnection) POST(ctx context.Context, uri string, payload string, handler func([]byte, error)) {
	if response, err := c._connection.Post(ctx, uri, message.AppJSON, bytes.NewReader([]byte(payload))); err == nil {
		if m, err := response.ReadBody(); err == nil {
			handler(m, ProcessMessageCode(response))
		} else {
			handler([]byte{}, err)
		}
	} else {
		log.WithFields(log.Fields{
			"Error": err.Error(),
		}).Error("Coap - POST")
	}
}

func (c *CoapDTLSConnection) AddToQueue(request CoapDTLSRequest) {
	c.mu.Lock()
	c.queue = append(c.queue, request)
	c.mu.Unlock()
}

func (c *CoapDTLSConnection) QueueLenght() int {
	c.mu.Lock()

	defer c.mu.Unlock()
	return len(c.queue)
}

func (c *CoapDTLSConnection) HandleQueue() {
	defer c.mu.Unlock()

	log.WithFields(log.Fields{
		"Items": len(c.queue),
	}).Debug("Tradfri: HandleQueue")

	var item CoapDTLSRequest
	c.mu.Lock()
	for len(c.queue) > 0 {
		item, c.queue = c.queue[0], c.queue[1:]
		switch item.RequestMethod {
		case "GET":
			ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
			c.GET(ctx, item.Uri, item.Handler)
		case "PUT":
			ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
			c.PUT(ctx, item.Uri, item.Payload, item.Handler)
		case "POST":
			ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
			c.POST(ctx, item.Uri, item.Payload, item.Handler)
		case "OBSERVE":
			c.Observe(item.Context, item.WaitGroup, item.Uri, item.Handler, item.KeepAlive)
		}
	}
}

func (c *CoapDTLSConnection) Observe(ctx context.Context, wg *sync.WaitGroup, uri string, handler func([]byte, error), keepAlive int) {
	var ticker *time.Ticker
	wg.Add(1)
	defer wg.Done()

	if e := ctx.Err(); e == context.Canceled {
		log.WithFields(log.Fields{
			"uri":   uri,
			"error": e.Error(),
		}).Debug("CoapDTLSConnection - Observe")
		return
	}

	if c._status != 2 {
		log.WithFields(log.Fields{
			"uri":   uri,
			"error": "Not connected to gateway, adding to queue",
		}).Debug("CoapDTLSConnection - Observe")
		c.HandleError(CoapDTLSRequest{Uri: uri, RequestMethod: "OBSERVE", Handler: handler, Context: ctx, WaitGroup: wg, KeepAlive: keepAlive})
		return
	}

	if keepAlive == 0 {
		ticker = time.NewTicker(1 * time.Second)
		ticker.Stop()
	} else {
		ticker = time.NewTicker(time.Duration(keepAlive) * time.Second)
		defer ticker.Stop()
	}

	for {
		_ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		obs, err := c._connection.Observe(_ctx, uri, func(req *pool.Message) {
			if m, err := req.ReadBody(); err == nil {
				handler(m, ProcessMessageCode(req))
			}
		})
		if err != nil {

			description := err.Error()

			if strings.Contains(description, "cannot write to connection") {
				c.HandleError(CoapDTLSRequest{Uri: uri, RequestMethod: "OBSERVE", Handler: handler, Context: ctx, WaitGroup: wg, KeepAlive: keepAlive})
			}

			log.WithFields(log.Fields{
				"uri":   uri,
				"Error": err.Error(),
			}).Error("CoapDTLSConnection - Observe - Init")
			return
		}

		select {
		case <-ticker.C:
			log.WithFields(log.Fields{
				"uri": uri,
			}).Debug("Observe keepalive")
			_ctx2, done2 := context.WithTimeout(context.Background(), 2*time.Second)
			defer done2()
			obs.Cancel(_ctx2)
			break
		case <-ctx.Done():
			log.WithFields(log.Fields{
				"uri": uri,
			}).Debug("Canceling observe")
			_ctx3, done3 := context.WithTimeout(context.Background(), 2*time.Second)
			defer done3()
			obs.Cancel(_ctx3)
			return
		}
	}
}
