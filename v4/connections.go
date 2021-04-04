package gocoap

import (
	piondtls "github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v2/dtls"
	"github.com/plgd-dev/go-coap/v2/udp/client"
)

var connection *client.ClientConn

func getDTLSConnection(param RequestParams) (*client.ClientConn, error) {
	if connection != nil {
		return connection, nil
	}

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
	connection = co
	return co, err
}

// CloseDTLSConnection closes the connection
func CloseDTLSConnection() error {
	if connection != nil {
		err := connection.Close()
		return err
	}
	return nil
}
