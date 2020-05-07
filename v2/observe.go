package gocoap

import (
	"fmt"
	"log"
	"time"

	coap "github.com/dustin/go-coap"
	"github.com/moroen/dtls"
)

type ObserveParams struct {
	Host            string
	Port            int
	URI             []string
	ID              string
	Key             string
	Req             coap.Message
	RetryConnection bool
}

var _obsListener *dtls.Listener
var _obsPeer *dtls.Peer

func getObserveDTLSConnection(params ObserveParams) (*dtls.Listener, *dtls.Peer, error) {
	retryCount := 1

	if params.RetryConnection {
		retryCount = 3
	}

	if _obsListener == nil {
		mks := dtls.NewKeystoreInMemory()
		dtls.SetKeyStores([]dtls.Keystore{mks})
		mks.AddKey(params.ID, []byte(params.Key))

		newListner, err := dtls.NewUdpListener(":0", time.Second*900)
		if err != nil {
			log.Println("Error: New UdpListener")
			return nil, nil, ErrorHandshake
		}
		_obsListener = newListner
	}

	if _obsPeer == nil {
		peerParams := &dtls.PeerParams{
			Addr:             fmt.Sprintf("%s:%d", params.Host, params.Port),
			Identity:         params.ID,
			HandshakeTimeout: time.Second * 3}

		for retryCount > 0 {
			newPeer, err := _obsListener.AddPeerWithParams(peerParams)
			if err != nil {
				log.Printf(" Create peer retrycount: %d", retryCount)
				retryCount--
			} else {
				retryCount = 0
				newPeer.UseQueue(true)
				_obsPeer = newPeer
			}
		}
	}

	if _obsPeer != nil {
		return _obsListener, _obsPeer, nil
	} else {
		return nil, nil, ErrorHandshake
	}
}

func Observe(params ObserveParams, returnMsg chan []byte, stop chan bool, status chan error) error {

	listener, peer, err := getObserveDTLSConnection(params)

	if err != nil {
		log.Println("Error Handshake - Giving up")
		return ErrorHandshake
	}

	for i, uri := range params.URI {

		params.Req = coap.Message{
			Type:      coap.NonConfirmable,
			Code:      coap.GET,
			MessageID: uint16(i),
		}

		params.Req.AddOption(coap.Observe, 1)
		params.Req.SetPathString(uri)

		data, err := params.Req.MarshalBinary()
		if err != nil {
			return ErrorUnknownError
		}

		err = peer.Write(data)
		if err != nil {
			return ErrorWriteTimeout
		}
	}

	go func(returnMsg chan []byte, stop chan bool, status chan error) {
		for {
			select {
			case <-stop:
				// log.Println("Stop received")
				listener.Shutdown()
				close(returnMsg)
				return
			default:
				respData, err := peer.Read(10 * time.Second)
				if err != nil {
					if listener == nil {
						log.Println("Observe Listener nil, reopening")
						listener, peer, err = getObserveDTLSConnection(params)
					}
					// log.Println("Timeout")
					continue
				}

				msg, err := coap.ParseMessage(respData)
				if err != nil {
					status <- ErrorBadData
					close(returnMsg)
				}
				returnMsg <- msg.Payload
			}
		}
	}(returnMsg, stop, status)
	return nil
}
