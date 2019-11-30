package gocoap

import (
	"context"
	"fmt"
	"strings"
	"time"

	coap "github.com/go-ocf/go-coap"
	"github.com/go-ocf/go-coap/codes"
	"github.com/pion/dtls"
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

func _returnErrorFromCode(code codes.Code) error {
	switch code {

	case codes.MethodNotAllowed:
		return MethodNotAllowed
	case codes.NotFound:
		return UriNotFound
	case codes.Content:
		return nil
	case codes.Changed:
		return nil
	case codes.Created:
		return nil
	case codes.BadRequest:
		return BadRequest
	case codes.Unauthorized:
		return Unauthorized
	}
	return nil
}

func _getConnection(params RequestParams) (conn *coap.ClientConn, err error) {
	if params.Id != "" {
		conn, err = coap.DialDTLS("udp", fmt.Sprintf("%s:%d", params.Host, params.Port), &dtls.Config{
			PSK: func(hint []byte) ([]byte, error) {
				// fmt.Printf("Server's hint: %s \n", hint)
				return []byte(params.Key), nil
			},
			PSKIdentityHint: []byte(params.Id),
			CipherSuites:    []dtls.CipherSuiteID{dtls.TLS_PSK_WITH_AES_128_CCM_8},
		})
	} else {
		conn, err = coap.Dial("udp", fmt.Sprintf("%s:%d", params.Host, params.Port))
	}

	return conn, err
}

// API

// GetRequest sends a get
func GetRequest(params RequestParams) (response []byte, err error) {
	conn, err := _getConnection(params)
	if err != nil {
		return response, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := conn.GetWithContext(ctx, params.Uri)
	if err != nil {
		return response, err
	}

	return resp.Payload(), _returnErrorFromCode(resp.Code())
}

// PutRequest sends a default Put-request
func PutRequest(params RequestParams) (response []byte, err error) {
	conn, err := _getConnection(params)
	if err != nil {
		return response, err
	}

	r := strings.NewReader(params.Payload)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := conn.PutWithContext(ctx, params.Uri, coap.TextPlain, r)
	if err != nil {
		return response, err
	}

	return resp.Payload(), _returnErrorFromCode(resp.Code())
}

// PostRequest sends a default Post-request
func PostRequest(params RequestParams) (response []byte, err error) {
	conn, err := _getConnection(params)
	if err != nil {
		return response, err
	}

	r := strings.NewReader(params.Payload)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := conn.PostWithContext(ctx, params.Uri, coap.TextPlain, r)

	if err != nil {
		return response, err
	}

	return resp.Payload(), _returnErrorFromCode(resp.Code())
}
