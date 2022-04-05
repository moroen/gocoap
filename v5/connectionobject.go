package gocoap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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

// var _cancel func()
// var _ctx context.Context

type CoapDTLSConnection struct {
	mu                 sync.Mutex
	Host               string
	Port               int
	Ident              string
	Key                string
	UseQueue           bool
	RetryConnect       bool
	OnConnect          func()
	OnDisconnect       func()
	OnCanceled         func()
	OnConnectionFailed func()
	_connection        *client.ClientConn
	_status            int
	queue              []CoapDTLSRequest
	ConnectContext     context.Context
	ConnectCancel      func()
	ObserveContext     context.Context
	ObserveWaitGroup   sync.WaitGroup
	ObserveDone        func()
	KeepAlive          int
	DisconnectTimer    int
}

type CoapDTLSRequest struct {
	RequestMethod string
	Uri           string
	Payload       string
	Handler       func([]byte, error)
	Context       context.Context
	WaitGroup     *sync.WaitGroup
}

func (c *CoapDTLSConnection) _keepAlive() {
	ticker := time.NewTicker(time.Duration(c.KeepAlive) * time.Second)
	for {
		select {
		case <-ticker.C:
			c.Disconnect()
			time.Sleep(500 * time.Millisecond)
			c.Connect()
		case <-c.ConnectContext.Done():
			return
		}
	}
}

func (c *CoapDTLSConnection) Connect() error {
	if c._status > 0 {
		return nil
	}

	c._status = 1 // Connecting

	c.ConnectContext, c.ConnectCancel = context.WithCancel(context.Background())

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
			c.ObserveContext, c.ObserveDone = context.WithCancel(context.Background())

			if c.OnConnect != nil {
				c._status = 2
				c.OnConnect()
			}

			if c.UseQueue {
				c.HandleQueue()
			}

			if c.KeepAlive > 0 {
				c._keepAlive()
			}

			if c.DisconnectTimer > 0 {
				c.TimedDisconnect()
			}

			return nil
		} else {
			if c.OnConnectionFailed != nil {
				c._status = 1
				c.OnConnectionFailed()
			}

			if !c.RetryConnect {
				return errors.New("unable to connect")
			}
		}
		select {
		case <-ticker.C:
			break
		case <-c.ConnectContext.Done():
			if c.OnCanceled != nil {
				c.OnCanceled()
			}
			return ConnectionContextCanceled
		}
	}
}

func (c *CoapDTLSConnection) TimedDisconnect() error {
	t := time.NewTimer(time.Second * time.Duration(c.DisconnectTimer))
	<-t.C
	c.Disconnect()
	return nil
}

func (c *CoapDTLSConnection) Disconnect() error {
	log.Debug("Disconnecting")

	if c.ConnectCancel != nil {
		c.ConnectCancel()
	}

	if c.ObserveDone != nil {
		c.ObserveDone()
	}
	c.ObserveWaitGroup.Wait()

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
		c.AddToQueue(request)
	}
	if c._status == 2 {
		c.Disconnect()
	}
	go c.Connect()
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
	log.WithFields(log.Fields{
		"Uri":     uri,
		"Payload": payload,
	}).Debug("CoapDTLSConnection.PUT")

	if c._status != 2 {
		log.WithFields(log.Fields{
			"Error": "Not connected",
		}).Error("COAP - PUT")

		c.HandleError(CoapDTLSRequest{RequestMethod: "PUT", Uri: uri, Payload: payload, Handler: handler})

		return
	}

	if response, err := c._connection.Put(ctx, uri, message.AppJSON, bytes.NewReader([]byte(payload))); err == nil {
		if m, err := response.ReadBody(); err == nil {
			handler(m, ProcessMessageCode(response))
		} else {
			handler([]byte{}, err)
		}
	} else {
		log.WithFields(log.Fields{
			"Error": err.Error(),
		}).Error("Coap - PUT - error")
		c.HandleError(CoapDTLSRequest{RequestMethod: "PUT", Uri: uri, Payload: payload, Handler: handler})
		return
	}
}

func (c *CoapDTLSConnection) POST(ctx context.Context, uri string, payload string, handler func([]byte, error)) {
	log.WithFields(log.Fields{
		"Uri":              uri,
		"Payload":          payload,
		"ConnectionStatus": c._status,
	}).Debug("CoapDTLSConnection.POST")

	if c._status != 2 {
		log.WithFields(log.Fields{
			"Error": "Not connected",
		}).Error("COAP - POST")
		c.HandleError(CoapDTLSRequest{RequestMethod: "POST", Uri: uri, Payload: payload, Handler: handler})
		return
	}

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
		c.HandleError(CoapDTLSRequest{RequestMethod: "POST", Uri: uri, Payload: payload, Handler: handler})
	}
}

func (c *CoapDTLSConnection) AddToQueue(request CoapDTLSRequest) {
	c.mu.Lock()
	c.queue = append(c.queue, request)
	c.mu.Unlock()

	log.WithFields(log.Fields{
		"Uri":          request.Uri,
		"Queue length": c.QueueLenght(),
	}).Debug("Added request to queue")
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
			go c.GET(ctx, item.Uri, item.Handler)
		case "PUT":
			ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
			go c.PUT(ctx, item.Uri, item.Payload, item.Handler)
		case "POST":
			ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
			go c.POST(ctx, item.Uri, item.Payload, item.Handler)
		case "OBSERVE":
			// go c.Observe(item.Context, item.WaitGroup, item.Uri, item.Handler, item.KeepAlive)
			// go c.Observe(item.Uri, item.Handler)
		}
	}
}

// func (c *CoapDTLSConnection) Observe(ctx context.Context, wg *sync.WaitGroup, uri string, handler func([]byte, error), keepAlive int) {
func (c *CoapDTLSConnection) Observe(uri string, handler func([]byte, error)) {
	c.ObserveWaitGroup.Add(1)
	defer c.ObserveWaitGroup.Done()

	if e := c.ObserveContext.Err(); e == context.Canceled {
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
		c.HandleError(CoapDTLSRequest{Uri: uri, RequestMethod: "OBSERVE", Handler: handler})
		return
	}

	_ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	obs, err := c._connection.Observe(_ctx, uri, func(req *pool.Message) {
		if m, err := req.ReadBody(); err == nil {
			handler(m, ProcessMessageCode(req))
		}
	})

	if err != nil {

		c.HandleError(CoapDTLSRequest{Uri: uri, RequestMethod: "OBSERVE", Handler: handler})

		log.WithFields(log.Fields{
			"uri":   uri,
			"Error": err.Error(),
		}).Error("CoapDTLSConnection - Observe - Init")
		return
	}

	<-c.ObserveContext.Done()
	log.WithFields(log.Fields{
		"uri": uri,
	}).Debug("Canceling observe")
	_ctx3, done3 := context.WithTimeout(context.Background(), 2*time.Second)
	defer done3()
	obs.Cancel(_ctx3)
}
