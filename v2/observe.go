package gocoap

import (
	"fmt"
	"log"
	"time"

	coap "github.com/dustin/go-coap"
	"github.com/moroen/dtls"
)

type ObserveParams struct {
	Host string
	Port int
	Uri  []string
	Id   string
	Key  string
	Req  coap.Message
}

var _obsListener *dtls.Listener
var _obsPeer *dtls.Peer

func getObserveDTLSConnection(params ObserveParams) (*dtls.Listener, *dtls.Peer, error) {
	if _obsListener == nil {
		mks := dtls.NewKeystoreInMemory()
		dtls.SetKeyStores([]dtls.Keystore{mks})
		mks.AddKey(params.Id, []byte(params.Key))

		newListner, err := dtls.NewUdpListener(":0", time.Second*900)
		if err != nil {
			return nil, nil, ErrorHandshake
		}
		_obsListener = newListner
	}

	if _obsPeer == nil {
		peerParams := &dtls.PeerParams{
			Addr:             fmt.Sprintf("%s:%d", params.Host, params.Port),
			Identity:         params.Id,
			HandshakeTimeout: time.Second * 3}

		newPeer, err := _obsListener.AddPeerWithParams(peerParams)
		if err != nil {
			return nil, nil, ErrorHandshake
		}

		newPeer.UseQueue(true)
		_obsPeer = newPeer
	}

	return _obsListener, _obsPeer, nil
}

func Observe(params ObserveParams, returnMsg chan []byte, stop chan bool, status chan error) error {
	listener, peer, err := getObserveDTLSConnection(params)
	if err != nil {
		return ErrorHandshake
	}

	for i, uri := range params.Uri {

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
