package gocoap

import (
	"bytes"
	"context"
	"fmt"

	"time"

	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/udp/message/pool"
	log "github.com/sirupsen/logrus"
	// coap "github.com/dustin/go-coap"
	// "github.com/eriklupander/dtls"
	// "github.com/moroen/dtls"
)

func _processMessage(resp *pool.Message) error {
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

func (c *CoapDTLSConnection) GET(ctx context.Context, uri string, handler func([]byte, error)) {
	log.WithFields(log.Fields{
		"Uri": uri,
	}).Debug("CoapDTLSConnection.GET")
	if response, err := c._connection.Get(ctx, uri); err == nil {
		if m, err := response.ReadBody(); err == nil {
			handler(m, _processMessage(response))
		} else {
			handler([]byte{}, err)
		}
	} else {
		fmt.Println(err.Error())
	}
}

func (c *CoapDTLSConnection) PUT(ctx context.Context, uri string, payload string, handler func([]byte, error)) {
	if response, err := c._connection.Put(ctx, uri, message.AppJSON, bytes.NewReader([]byte(payload))); err == nil {
		if m, err := response.ReadBody(); err == nil {
			handler(m, _processMessage(response))
		} else {
			handler([]byte{}, err)
		}
	} else {
		fmt.Println(err.Error())
	}
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

func _requestDTLS(params RequestParams) (retmsg []byte, err error) {
	var response *pool.Message

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	co, err := getDTLSConnection(ctx, params)
	if err != nil {
		return nil, err
	}

	path := params.Uri

	if params.Method == GET {
		if resp, err := co.Get(ctx, path); err == nil {
			response = resp
		}
	}

	if params.Method == PUT {
		payload := bytes.NewReader([]byte(params.Payload))
		if resp, err := co.Put(ctx, path, message.AppJSON, payload); err == nil {
			response = resp
		}
	}

	if params.Method == POST {
		payload := bytes.NewReader([]byte(params.Payload))
		if resp, err := co.Post(ctx, path, message.AppJSON, payload); err == nil {
			response = resp
		}
	}

	if err == nil {
		if response == nil {
			log.WithFields(log.Fields{
				"path": path,
			}).Debug("Response is nil")

			CloseDTLSConnection()
			return _requestDTLS(params)
		} else {
			m, err := response.ReadBody()
			if err != nil {
				return nil, err
			}

			err = _processMessage(response)
			return m, err
		}
	} else {
		return nil, err
	}
}

func SetRetry(limit uint, delay int) {
	_retryLimit = limit
	_retryDelay = delay
}

func GetRequestWithContext(ctx context.Context, params RequestParams, retrydelay int) (response []byte, err error) {
	ticker := time.NewTicker(time.Duration(retrydelay) * time.Second)
	for {
		if res, err := GetRequest(params); err == nil {
			return res, err
		} else {
			log.WithFields(log.Fields{
				"URL":   params.Uri,
				"Error": err,
			}).Error("GetRequestWithContext")
		}
		select {
		case <-ticker.C:
			break
		case <-ctx.Done():
			fmt.Println("tock")
			return nil, ErrorHandshake
		}
	}
}

// GetRequest sends a default get
func GetRequest(params RequestParams) (response []byte, err error) {
	var msg []byte

	params.Method = GET

	if params.Id != "" {
		msg, err = _requestDTLS(params)
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
		msg, err = _requestDTLS(params)
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
		msg, err = _requestDTLS(params)
	} else {
		msg, err = _request(params)
	}

	return msg, err
}
