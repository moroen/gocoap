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

//export sum
func sum(self, args *C.PyObject) *C.PyObject {
	var a, b C.longlong
	if C.PyArg_ParseTuple_LL(args, &a, &b) == 0 {
		return nil
	}
	return C.PyLong_FromLongLong(a + b)
}

/*
// export coapSetGateway
func coapSetGateway(ip, ident, psk *C.char) int {
	fmt.Println(C.GoString(ip))
}*/

/*
func coapSetGateway(self, args *C.PyObject) *C.PyObject {
}*/

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

func main() {}
