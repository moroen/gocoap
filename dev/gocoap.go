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

var _listner *dtls.Listener
var _peer *dtls.Peer

var retryLimit = 3

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
	if _listner == nil {
		mks := dtls.NewKeystoreInMemory()
		dtls.SetKeyStores([]dtls.Keystore{mks})
		mks.AddKey(params.Id, []byte(params.Key))

		newListner, err := dtls.NewUdpListener(":0", time.Second*900)
		if err != nil {
			return nil, nil, ErrorHandshake
		}
		_listner = newListner
	}

	if _peer == nil {
		peerParams := &dtls.PeerParams{
			Addr:             fmt.Sprintf("%s:%d", params.Host, params.Port),
			Identity:         params.Id,
			HandshakeTimeout: time.Second * 3}

		newPeer, err := _listner.AddPeerWithParams(peerParams)
		if err != nil {
			return nil, nil, ErrorHandshake
		}

		newPeer.UseQueue(true)
		_peer = newPeer
	}

	return _listner, _peer, nil
}

func _requestDTLS(params RequestParams, retry int) (retmsg coap.Message, err error) {

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
		_listner = nil

		if retry < retryLimit {
			log.Println("Retrying Write request")
			return _requestDTLS(params, retry+1)
		} else {
			return coap.Message{}, err
		}
	}

	respData, err := peer.Read(time.Second)
	if err != nil {
		log.Println("Read Timeout")

		listner.Shutdown()
		_peer = nil
		_listner = nil

		if retry < retryLimit {
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

	err = _processMessage(msg)
	return msg, err
}

// Observe a uri
func Observe(params RequestParams, returnMsg chan []byte, stop chan bool) {
	log.Println("Observing ", params.Uri)

	listner, peer, err := getDTLSConnection(params)
	if err != nil {
		panic(err.Error())
	}

	params.Req = coap.Message{
		Type:      coap.NonConfirmable,
		Code:      coap.GET,
		MessageID: 12345,
	}

	params.Req.AddOption(coap.Observe, 1)
	params.Req.SetPathString(params.Uri)

	data, err := params.Req.MarshalBinary()
	if err != nil {
		panic(err.Error())
	}

	err = peer.Write(data)
	if err != nil {
		panic(err.Error())
	}

	params.Req = coap.Message{
		Type:      coap.NonConfirmable,
		Code:      coap.GET,
		MessageID: 12346,
	}

	params.Req.AddOption(coap.Observe, 1)
	params.Req.SetPathString("/15001/65550")

	data, err = params.Req.MarshalBinary()
	if err != nil {
		panic(err.Error())
	}

	err = peer.Write(data)
	if err != nil {
		panic(err.Error())
	}

	for {
		select {
		case <-stop:
			log.Println("Stop received")
			listner.Shutdown()
			return
		default:
			respData, err := peer.Read(10 * time.Second)
			if err != nil {
				// log.Println("Timeout")
				continue
			}

			msg, err := coap.ParseMessage(respData)
			if err != nil {
				panic(err.Error())
			}
			returnMsg <- msg.Payload
		}
	}
}

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
