#!/usr/bin/env python3

import pycoap as coap


def getCoreDTLS():
    # With DTLS    
    coap.SetGateway("localhost:5684","USER_ID","PSK")
    result = coap.Request("/.well-known/core")
    print(result)

def getCore():
    # Without DTLS    
    coap.SetGateway("localhost:5683","","")
    result = coap.Request("/.well-known/core")
    print(result)

def sendPayload():
    coap.SetGateway("localhost:5684","USER_ID","PSK")
    print(coap.PutRequest("15001/65540", "{ \"3311\": [{ \"5850\": 1 }] }"))

if __name__ == '__main__':
    getCore()    
