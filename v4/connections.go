package gocoap

import (
	piondtls "github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v2/dtls"
	"github.com/plgd-dev/go-coap/v2/udp/client"
	log "github.com/sirupsen/logrus"
)

var _connection *client.ClientConn
var _retryLimit uint = 3
var _retryDelay = 1

func reconnectDtlsConnection(param RequestParams) (*client.ClientConn, error) {
	log.Debug("reconnectDtlsConnection")
	CloseDTLSConnection()
	conn, err := getDTLSConnection(param)
	return conn, err
}

func getDTLSConnection(param RequestParams) (*client.ClientConn, error) {
	if _connection != nil {
		// log.Debug("getDTLSConnection: Using old connection")
		return _connection, nil
	}

	// log.Debug("getDTLSConnection: Creating new connection")

	if conn, err := createDTLSConnection(param); err == nil {
		_connection = conn
		return _connection, nil
	} else {
		_connection = nil
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("getDTLSConnection")
		err = ErrorHandshake
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
		// log.Println("Connection closing")
		err := _connection.Close()
		if err != nil {
			return err
		}
		_connection = nil
	}
	return nil
}
