package gocoap

import (
	"fmt"
	"log"

	"time"

	coap "github.com/dustin/go-coap"
	// "github.com/eriklupander/dtls"
	"github.com/moroen/dtls"
)

type RequestParams struct {
	Host    string
	Port    int
	Uri     string
	Id      string
	Key     string
	Req     coap.Message
	Payload string
}

var _listener *dtls.Listener
var _peer *dtls.Peer

var _retryLimit = 3

func _processMessage(msg coap.Message) error {
	switch msg.Code {
	case coap.MethodNotAllowed:
		return MethodNotAllowed
	case coap.NotFound:
		return UriNotFound
	case coap.Content:
		return nil
	case coap.Changed:
		return nil
	case coap.Created:
		return nil
	case coap.BadRequest:
		return BadRequest
	case coap.Unauthorized:
		return Unauthorized
	}

	return ErrorUnknownError
}

func _request(params RequestParams) (retmsg coap.Message, err error) {
	conn, err := coap.Dial("udp", fmt.Sprintf("%s:%d", params.Host, params.Port))
	if err != nil {
		return retmsg, err
	}

	resp, err := conn.Send(params.Req)
	if err != nil {
		return retmsg, err
	}

	err = _processMessage(*resp)

	return *resp, err
}

func getDTLSConnection(params RequestParams) (*dtls.Listener, *dtls.Peer, error) {
	if _listener == nil {
		for i := 0; i < _retryLimit; i++ {
			fmt.Printf("Creating new listener: try %d\n", i)
			mks := dtls.NewKeystoreInMemory()
			dtls.SetKeyStores([]dtls.Keystore{mks})
			mks.AddKey(params.Id, []byte(params.Key))

			newListner, err := dtls.NewUdpListener(":0", time.Second*2)
			if err != nil {
				fmt.Print("listener failed, retry")
			} else {
				_listener = newListner
				break
			}
		}
		if _listener == nil {
			return nil, nil, ErrorHandshake
		}
	}

	if _peer == nil {
		for i := 0; i < _retryLimit; i++ {

			fmt.Printf("Creating new peer: try %d\n", i)

			peerParams := &dtls.PeerParams{
				Addr:             fmt.Sprintf("%s:%d", params.Host, params.Port),
				Identity:         params.Id,
				HandshakeTimeout: time.Second * 2}

			newPeer, err := _listener.AddPeerWithParams(peerParams)
			if err != nil {
				fmt.Print("peer failed, retry")
			} else {
				newPeer.UseQueue(true)
				_peer = newPeer
				return _listener, _peer, nil
			}
		}
		return nil, nil, ErrorHandshake
	}

	return _listener, _peer, nil
}

func SetRetryLimit(limit int) {
	_retryLimit = limit
}

func CloseDTLSConnection() error {
	fmt.Println("Closing connection")
	if _listener != nil {
		_listener.Shutdown()
	}

	_listener = nil
	_peer = nil

	return nil
}

func _requestDTLS(params RequestParams, retry int) (retmsg coap.Message, err error) {

	/*
		if params.Req.Code == coap.PUT {
			fmt.Println("Doing a put request")
			// listner.Shutdown()
			_peer = nil
			_listener = nil
		}
	*/

	fmt.Println("_requestDTLS called")

	listner, peer, err := getDTLSConnection(params)
	if err != nil {
		return coap.Message{}, err
	}

	data, err := params.Req.MarshalBinary()
	if err != nil {
		return coap.Message{}, ErrorUnknownError
	}

	err = peer.Write(data)
	if err != nil {
		log.Println("Read Timeout")

		listner.Shutdown()
		_peer = nil
		_listener = nil

		if retry < _retryLimit {
			log.Println("Retrying Write request")
			return _requestDTLS(params, retry+1)
		}
		return coap.Message{}, err

	}

	respData, err := peer.Read(time.Second)
	if err != nil {
		log.Println("Read Timeout")

		listner.Shutdown()
		_peer = nil
		_listener = nil

		if retry < _retryLimit {
			log.Println("Retrying Read request")
			return _requestDTLS(params, retry+1)
		} else {
			return coap.Message{}, err
		}
	}

	msg, err := coap.ParseMessage(respData)
	if err != nil {
		return coap.Message{}, ErrorBadData
	}

	// fmt.Println(msg.Code)

	if msg.Code == coap.Changed {
		params.Req.Code = coap.GET
		params.Req.Payload = nil
		msg, err = _requestDTLS(params, 0)
	}

	err = _processMessage(msg)
	fmt.Println("_requestDTLS done")
	return msg, err
}

// Observe a uri

// GetRequest sends a default get
func GetRequest(params RequestParams) (response []byte, err error) {
	params.Req = coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.GET,
		MessageID: 1,
	}

	params.Req.SetPathString(params.Uri)

	var msg coap.Message

	if params.Id != "" {
		msg, err = _requestDTLS(params, 0)
	} else {
		msg, err = _request(params)
	}
	return msg.Payload, err
}

// PutRequest sends a default Put-request
func PutRequest(params RequestParams) (response []byte, err error) {

	params.Req = coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.PUT,
		MessageID: 1,
		Payload:   []byte(params.Payload),
	}

	params.Req.SetPathString(params.Uri)

	var msg coap.Message

	if params.Id != "" {
		msg, err = _requestDTLS(params, 0)
	} else {
		msg, err = _request(params)
	}

	return msg.Payload, err
}

// PostRequest sends a default Post-request
func PostRequest(params RequestParams) (response []byte, err error) {
	params.Req = coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.POST,
		MessageID: 1,
		Payload:   []byte(params.Payload),
	}

	params.Req.SetPathString(params.Uri)

	var msg coap.Message

	if params.Id != "" {
		msg, err = _requestDTLS(params, 0)
	} else {
		msg, err = _request(params)
	}

	return msg.Payload, err
}
