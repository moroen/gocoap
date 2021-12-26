package gocoap

import (
	"context"
	"fmt"
	"sync"
	"time"

	piondtls "github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v2/dtls"
	"github.com/plgd-dev/go-coap/v2/udp/client"
	log "github.com/sirupsen/logrus"
)

var _connection *client.ClientConn
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

/*
func reconnectDtlsConnection(param RequestParams) (*client.ClientConn, error) {
	log.Debug("reconnectDtlsConnection")
	CloseDTLSConnection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_cancel = cancel
	conn, err := getDTLSConnection(ctx, param)

	return conn, err
}

func getDTLSConnection(ctx context.Context, param RequestParams) (*client.ClientConn, error) {

	if _connection != nil {
		log.Debug("getDTLSConnection: Using old connection")
		return _connection, nil
	}

	if conn, err := createDTLSConnection(param); err == nil {
		log.Info(fmt.Sprintf("Connected to tradfri at %s:5684", param.Host))
		_connection = conn
		return _connection, nil
	} else {
		return nil, err
	}
}

func createDTLSConnection(param RequestParams) (*client.ClientConn, error) {
	co, err := dtls.Dial(param.getHost(), &piondtls.Config{
		PSK: func(hint []byte) ([]byte, error) {
			// fmt.Printf("Server's hint: %s \n", hint)
			return []byte(param.Key), nil
		},
		PSKIdentityHint: []byte(param.Id),
		CipherSuites:    []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
	})
	if err != nil {
		err = ErrorHandshake
	}
	_connection = co
	return co, err
}

// CloseDTLSConnection closes the connection
func CloseDTLSConnection() error {
	if _connection != nil {
		err := _connection.Close()
		if err != nil {
			return err
		}
		_connection = nil
	}
	return nil
}
*/
