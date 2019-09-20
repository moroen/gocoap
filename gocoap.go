package gocoap

import (
	"errors"
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

// ErrorTimeout error
var ErrorTimeout = errors.New("COAP Error: Connection timeout")

// ErrorBadIdent error
var ErrorBadIdent = errors.New("COAP DTLS Error: Wrong credentials?")

// ErrorNoConfig error
var ErrorNoConfig = errors.New("COAP Error: No config")

/*
func _getRequest(URI string, c chan CoapResult) {

	var result CoapResult

	conf, err := GetConfig()
	if err != nil {
		result.err = ErrorNoConfig
		c <- result
		return
	}

	conn, err := canopus.DialDTLS(conf.Gateway, conf.Identity, conf.Passkey)
	if err != nil {
		result.err = err
		c <- result
		return
	}

	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Get)
	req.SetStringPayload("Hello, canopus")
	req.SetRequestURI(URI)

	resp, err := conn.Send(req)
	if err != nil {
		result.err = ErrorBadIdent
		c <- result
		return
	}

	// response := resp.GetMessage().GetPayload()
	result.err = nil
	result.msg = resp.GetMessage().GetPayload()
	c <- result
}
*/

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
		HandshakeTimeout: time.Second * 15}

	peer, err := listner.AddPeerWithParams(peerParams)
	if err != nil {
		panic(err.Error())
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

	return msg, nil
}

func _getRequest(params RequestParams, c chan coap.Message) {
	params.Req = coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.GET,
		MessageID: 1,
	}

	params.Req.SetPathString(params.Uri)

	var msg coap.Message
	var err error

	if params.Id != "" {
		msg, err = _requestDTLS(params)
	} else {
		msg, err = _request(params)
	}

	if err != nil {
		panic(err.Error())
	}
	c <- msg
}

func _putRequest(params RequestParams, c chan coap.Message) {
	params.Req = coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.PUT,
		MessageID: 1,
		Payload:   []byte(params.Payload),
	}
	params.Req.SetPathString(params.Uri)

	var msg coap.Message
	var err error

	if params.Id != "" {
		msg, err = _requestDTLS(params)
	} else {
		msg, err = _request(params)
	}
	if err != nil {
		panic(err.Error())
	}
	c <- msg
}

// GetRequest sends a default get
func GetRequest(params RequestParams) (response []byte, err error) {
	c := make(chan coap.Message)

	go _getRequest(params, c)

	select {
	case res := <-c:
		return res.Payload, nil
	case <-time.After(time.Second * 60):
		return nil, ErrorTimeout
	}
}

// PutRequest sends a default Put-request
func PutRequest(params RequestParams) (response []byte, err error) {
	c := make(chan coap.Message)

	go _putRequest(params, c)

	select {
	case res := <-c:
		return res.Payload, nil
	case <-time.After(time.Second * 5):
		return nil, ErrorTimeout
	}
}
