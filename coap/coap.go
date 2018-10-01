package main

// #cgo pkg-config: python3
// #define Py_LIMITED_API
// #include <Python.h>
// int PyArg_ParseTuple_LL(PyObject *, long long *, long long *);
// int PyArg_ParseTuple_S(PyObject *, char *);
// char * ParseStringArgument(PyObject *);
import "C"

import (
	"errors"
	"time"

	"github.com/moroen/canopus"
)

type CoapRequest struct {
	Gateway  string
	Uri      string
	Identity string
	Passkey  string
	Payload  string
}

type CoapResult struct {
	msg canopus.MessagePayload
	err error
}

var ErrorTimeout = errors.New("COAP Error: Connection timeout")
var ErrorBadIdent = errors.New("COAP DTLS Error: Wrong credentials?")
var ErrorNoConfig = errors.New("COAP Error: No config")

func _getRequest(request CoapRequest, c chan CoapResult) {

	var result CoapResult
	var conn canopus.Connection
	var err error

	if request.Identity != "" {
		conn, err = canopus.DialDTLS(request.Gateway, request.Identity, request.Passkey)
	} else {
		conn, err = canopus.Dial(request.Gateway)
	}

	if err != nil {
		result.err = err
		c <- result
		return
	}

	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Get)
	req.SetStringPayload("Hello, canopus")
	req.SetRequestURI(request.Uri)

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

func _putRequest(request CoapRequest, c chan CoapResult) {
	var result CoapResult

	var conn canopus.Connection
	var err error

	if request.Identity != "" {
		conn, err = canopus.DialDTLS(request.Gateway, request.Identity, request.Passkey)
	} else {
		conn, err = canopus.Dial(request.Gateway)

	}

	if err != nil {
		result.err = err
		c <- result
		return
	}

	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Put)
	req.SetRequestURI(request.Uri)
	req.SetStringPayload(request.Payload)

	resp, err := conn.Send(req)
	if err != nil {
		result.err = ErrorBadIdent
		c <- result
		return
	}

	result.msg = resp.GetMessage().GetPayload()
	result.err = nil
	c <- result
}

// Export GetRequestDTLS
func GetRequestDTLS(gateway, uri, ident, key string) (msg canopus.MessagePayload, err error) {
	c := make(chan CoapResult)

	req := CoapRequest{Gateway: gateway, Uri: uri, Identity: ident, Passkey: key, Payload: ""}

	go _getRequest(req, c)

	select {
	case res := <-c:
		return res.msg, res.err
	case <-time.After(time.Second * 5):
		return nil, ErrorTimeout
	}
}

// Export GetRequest
func GetRequest(gateway, uri string) (msg canopus.MessagePayload, err error) {
	return GetRequestDTLS(gateway, uri, "", "")
}

func PutRequestDTLS(gateway, uri, ident, key, payload string) (msg canopus.MessagePayload, err error) {
	c := make(chan CoapResult)

	req := CoapRequest{Gateway: gateway, Uri: uri, Identity: ident, Passkey: key, Payload: payload}

	go _putRequest(req, c)

	select {
	case _ = <-c:
		return GetRequestDTLS(gateway, uri, ident, key)
	case <-time.After(time.Second * 5):
		return nil, ErrorTimeout
	}
}

func PutRequest(gateway, uri, payload string) (msg canopus.MessagePayload, err error) {
	return PutRequestDTLS(gateway, uri, "", "", payload)
}

/* func Observe(URI string) {

	conf, err := GetConfig()
	if err != nil {
		result.err = ErrorNoConfig
		c <- result
		return
	}

	conn, err := canopus.DialDTLS(conf.Gateway, conf.Identity, conf.Passkey)

	tok, err := conn.ObserveResource("/15001/65540")
	if err != nil {
		panic(err.Error())
	}

	obsChannel := make(chan canopus.ObserveMessage)
	done := make(chan bool)
	go conn.Observe(obsChannel)

	notifyCount := 0
	go func() {
		for {
			select {
			case obsMsg, open := <-obsChannel:
				if open {
					if notifyCount == 5 {
						fmt.Println("[CLIENT >> ] Canceling observe after 5 notifications..")
						go conn.CancelObserveResource("watch/this", tok)
						go conn.StopObserve(obsChannel)
						done <- true
						return
					} else {
						notifyCount++
						// msg := obsMsg.Msg\
						resource := obsMsg.GetResource()
						val := obsMsg.GetValue()

						fmt.Println("[CLIENT >> ] Got Change Notification for resource and value: ", notifyCount, resource, val)
					}
				} else {
					done <- true
					return
				}
			}
		}
	}()
	<-done
	fmt.Println("Done")
}
*/
func main() {}
