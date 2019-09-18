package main

// #cgo pkg-config: python3
// #define Py_LIMITED_API
// #include <Python.h>
// int PyArg_ParseTuple_LL(PyObject *, long long *, long long *);
// int PyArg_ParseTuple_S(PyObject *, char *);
// char * ParseStringArgument(PyObject *);
import "C"
import "fmt"

// Python Functions

//export coapRequest
func coapRequest(gateway, uri *C.char) *C.char {

	msg, err := GetRequest(C.GoString(gateway), C.GoString(uri))
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	// return C.PyUnicode_FromString(C.CString(msg.String()))
	return C.CString(msg.String())
}

//export coapRequestDTLS
func coapRequestDTLS(gateway, uri, ident, key *C.char) *C.char {

	msg, err := GetRequestDTLS(C.GoString(gateway), C.GoString(uri), C.GoString(ident), C.GoString(key))

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	// return C.PyUnicode_FromString(C.CString(msg.String()))
	return C.CString(msg.String())
}

//export coapPutRequest
func coapPutRequest(gateway, uri, payload *C.char) *C.char {
	msg, err := PutRequest(C.GoString(gateway), C.GoString(uri), C.GoString(payload))
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return C.CString(msg.String())
}

//export coapPutRequestDTLS
func coapPutRequestDTLS(gateway, uri, ident, key, payload *C.char) *C.char {
	msg, err := PutRequestDTLS(C.GoString(gateway), C.GoString(uri), C.GoString(ident), C.GoString(key), C.GoString(payload))
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return C.CString(msg.String())
}
