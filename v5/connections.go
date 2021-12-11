package gocoap

import (
	"context"
	"fmt"
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

type CoapDTLSConnection struct {
	Host               string
	Port               int
	Ident              string
	Key                string
	OnConnect          func()
	OnDisconnect       func()
	OnCanceled         func()
	OnConnectionFailed func()
	_connection        *client.ClientConn
}

func (c *CoapDTLSConnection) Connect(ctx context.Context) error {
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
				c.OnConnect()
			}
			return nil
		} else {
			if c.OnConnectionFailed != nil {
				c.OnConnectionFailed()
			}
		}
		select {
		case <-ticker.C:
			break
		case <-ctx.Done():
			if c.OnCanceled != nil {
				c.OnCanceled()
			}
			return ConnectionContextCanceled
		}
	}
}

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
