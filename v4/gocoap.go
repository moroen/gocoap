package gocoap

import (
	"bytes"
	"context"
	"log"

	"time"

	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/udp/message/pool"
	// coap "github.com/dustin/go-coap"
	// "github.com/eriklupander/dtls"
	// "github.com/moroen/dtls"
)

// var _listener *dtls.Listener
// var _peer *dtls.Peer

var _retryLimit = 3

func _processMessage(resp *pool.Message) error {
	log.Printf("%+v", resp)
	log.Println(resp.Code())
	switch resp.Code() {
	case codes.Content:
		return nil
	case codes.MethodNotAllowed:
		return MethodNotAllowed
	case codes.NotFound:
		return UriNotFound
	case codes.Changed:
		return nil
	case codes.Created:
		return nil
	case codes.BadRequest:
		return BadRequest
	case codes.Unauthorized:
		return Unauthorized
	}

	return ErrorUnknownError
}

func _request(params RequestParams) (retmsg []byte, err error) {
	/*
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
	*/
	return nil, nil
}

// SetRetryLimit sets number of retries, default i 3
func SetRetryLimit(limit int) {
	_retryLimit = limit
}

func _requestDTLS(params RequestParams, retry int) (retmsg []byte, err error) {

	co, err := getDTLSConnection(params)
	if err != nil {
		return nil, err
	}

	path := params.Uri

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if params.Method == GET {
		resp, err := co.Get(ctx, path)
		if err != nil {
			log.Fatalf("Error sending request: %v", err)
		}

		m, err := resp.ReadBody()
		if err != nil {
			return nil, err
		}
		// log.Printf("Response payload: %v", string(m))

		err = _processMessage(resp)
		return m, err
	}

	if params.Method == PUT {
		payload := bytes.NewReader([]byte(params.Payload))

		resp, err := co.Put(ctx, path, message.AppJSON, payload)
		if err != nil {
			log.Fatal(err.Error())
		}

		m, err := resp.ReadBody()
		if err != nil {
			return nil, err
		}

		return m, _processMessage(resp)
	}

	if params.Method == POST {
		payload := bytes.NewReader([]byte(params.Payload))

		resp, err := co.Post(ctx, params.Uri, message.AppJSON, payload)
		if err != nil {
			return nil, err
		}

		m, err := resp.ReadBody()
		if err != nil {
			return nil, err
		}

		return m, _processMessage(resp)
	}

	return nil, nil
}

// GetRequest sends a default get
func GetRequest(params RequestParams) (response []byte, err error) {
	var msg []byte

	params.Method = GET

	if params.Id != "" {
		msg, err = _requestDTLS(params, 0)
	} else {
		msg, err = _request(params)
	}

	return msg, err
}

// PutRequest sends a default Put-request
func PutRequest(params RequestParams) (response []byte, err error) {
	var msg []byte

	if params.Payload == "" {
		return nil, ErrorNoPayload
	}

	params.Method = PUT

	if params.Id != "" {
		msg, err = _requestDTLS(params, 0)
	} else {
		msg, err = _request(params)
	}

	return msg, err
}

// PostRequest sends a default Post-request
func PostRequest(params RequestParams) (response []byte, err error) {
	var msg []byte

	if params.Payload == "" {
		return nil, ErrorNoPayload
	}

	params.Method = POST

	if params.Id != "" {
		msg, err = _requestDTLS(params, 0)
	} else {
		msg, err = _request(params)
	}

	return msg, err
}
