package main

// #cgo pkg-config: python3
// #define Py_LIMITED_API
// #include <Python.h>
// int PyArg_ParseTuple_LL(PyObject *, long long *, long long *);
// int PyArg_ParseTuple_S(PyObject *, char *);
// char * ParseStringArgument(PyObject *);
import "C"

import (
	"fmt"

	coap "github.com/moroen/go-tradfricoap"
)

//export coapSetGateway
func coapSetGateway(ip, ident, key *C.char) C.int {
	conf := coap.GatewayConfig{
		Gateway:  C.GoString(ip),
		Identity: C.GoString(ident),
		Passkey:  C.GoString(key),
	}

	coap.SetConfig(conf)
	return 1
}

//export coapRequest
func coapRequest(uri *C.char) *C.char {

	s := C.GoString(uri)
	// fmt.Println(C.GoString(s))

	msg, err := coap.GetRequest(s)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	// return C.PyUnicode_FromString(C.CString(msg.String()))
	return C.CString(msg.String())
}

//export coapPutRequest
func coapPutRequest(uri, payload *C.char) *C.char {
	sURI := C.GoString(uri)
	sPayLoad := C.GoString(payload)

	msg, err := coap.PutRequest(sURI, sPayLoad)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return C.CString(msg.String())
}

func main() {}
