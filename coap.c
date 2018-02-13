#define Py_LIMITED_API
#include <Python.h>
#include <stdio.h>

PyObject * sum(PyObject *, PyObject *);
// PyObject * coapRequest(PyObject *, PyObject *);
int coapSetGateway(char*, char*, char*);
char * coapRequest(char *);

// Workaround missing variadic function support
// https://github.com/golang/go/issues/975
int PyArg_ParseTuple_LL(PyObject * args, long long * a, long long * b) {  
    return PyArg_ParseTuple(args, "LL", a, b);
}

int PyArg_ParseTuple_SSS(PyObject * args, char * a, char *b, char *c) {
    return PyArg_ParseTuple(args, "sss", a,b,c);
}

const char * ParseStringArgument(PyObject * args) {
    const char* s;

    if (!PyArg_ParseTuple(args, "s", &s))
        return NULL;

    return s;
}

/*
static PyObject * test(PyObject *self, PyObject *args)  
{
    const long long a, b;
    const char* s;

    if (!PyArg_ParseTuple(args, "s", &s))
        return NULL;

    printf("%s", s);

    Py_RETURN_NONE; 
}
*/

PyObject * setGateway(PyObject *self, PyObject *args) {
    char *ip, *ident, *psk;

    if (!PyArg_ParseTuple(args, "sss", &ip, &ident, &psk))
        Py_RETURN_NONE;

    coapSetGateway(ip, ident, psk);
    Py_RETURN_NONE;
}

PyObject * request(PyObject *self, PyObject *args) {
    char *uri, *res;

    if (!PyArg_ParseTuple(args, "s", &uri))
        Py_RETURN_NONE;
    
    res = coapRequest(uri);
    if (!res)
        Py_RETURN_NONE;
        
    return PyUnicode_FromString(res);
}

static PyMethodDef CoapMethods[] = {  
    // {"test", test, METH_VARARGS, "Add two numbers."},
    {"sum", sum, METH_VARARGS, "Add two numbers."},
    {"SetGateway", setGateway, METH_VARARGS, "Set gateway info"},
    {"Request", request, METH_VARARGS, "Make a COAP Request."},
    {NULL, NULL, 0, NULL}
};

static struct PyModuleDef coapmodule = {  
   PyModuleDef_HEAD_INIT, "coap", NULL, -1, CoapMethods
};

PyMODINIT_FUNC  
PyInit_coap(void)  
{
    return PyModule_Create(&coapmodule);
}