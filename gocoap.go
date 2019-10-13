package gocoap

import (
	"fmt"

	"time"

	coap "github.com/dustin/go-coap"
	"github.com/eriklupander/dtls"
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

func _request(params RequestParams) (retmsg coap.Message, err error) {
	return params.Req, nil
}

func _requestDTLS(params RequestParams) (retmsg coap.Message, err error) {
	mks := dtls.NewKeystoreInMemory()
	dtls.SetKeyStores([]dtls.Keystore{mks})
	mks.AddKey(params.Id, []byte(params.Key))

	listner, err := dtls.NewUdpListener(":0", time.Second*900)
	if err != nil {
		panic(err.Error())
	}

	peerParams := &dtls.PeerParams{
		Addr:             fmt.Sprintf("%s:%d", params.Host, params.Port),
		Identity:         params.Id,
		HandshakeTimeout: time.Second * 3}

	peer, err := listner.AddPeerWithParams(peerParams)
	if err != nil {
		return coap.Message{}, ErrorHandshake
	}

	peer.UseQueue(true)

	data, err := params.Req.MarshalBinary()
	if err != nil {
		panic(err.Error())
	}

	err = peer.Write(data)
	if err != nil {
		panic(err.Error())
	}

	respData, err := peer.Read(time.Second)
	if err != nil {
		panic(err.Error())
	}

	msg, err := coap.ParseMessage(respData)
	if err != nil {
		panic(err.Error())
	}

	err = listner.Shutdown()
	if err != nil {
		panic(err.Error())
	}

	switch msg.Code {
	case coap.MethodNotAllowed:
		return msg, MethodNotAllowed
	case coap.NotFound:
		return msg, UriNotFound
	case coap.Content:
		return msg, nil
	case coap.Changed:
		return msg, nil
	case coap.Created:
		return msg, nil
	case coap.BadRequest:
		return msg, BadRequest
	case coap.Unauthorized:
		return msg, Unauthorized
	}

	return msg, UnknownError
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
		msg, err = _requestDTLS(params)
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
		msg, err = _requestDTLS(params)
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
		msg, err = _requestDTLS(params)
	} else {
		msg, err = _request(params)
	}

	return msg.Payload, err
}