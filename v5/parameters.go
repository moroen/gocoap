package gocoap

import "fmt"

type RequestMethod int

var (
	GET  RequestMethod = 1
	PUT  RequestMethod = 2
	POST RequestMethod = 3
)

type RequestParams struct {
	Host    string
	Port    int
	Uri     string
	Id      string
	Key     string
	Payload string
	Method  RequestMethod
}

type ObserveParams struct {
	Host            string
	Port            int
	Uri             []string
	Id              string
	Key             string
	RetryConnection bool
}

func (r RequestParams) getHost() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
