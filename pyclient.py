#!/usr/bin/env python3

import pycoap
import argparse

parser = argparse.ArgumentParser()
parser.add_argument("uri")
parser.add_argument("payload", nargs="?")
parser.add_argument("--ident")
parser.add_argument("--key")

class MalformedURI(Exception):
    pass

class MissingCredentials(Exception):
    pass

def getCoreDTLS():
    # With DTLS    
    result =pycoap.DTLSRequest("192.168.1.15:5684", "/.well-known/core", "44d68a62-e6d5-4743-a4fb-ba1317c0e7a5", "eQrKSrpoWgdOPIbw")
    print(result)

def getCore():
    # Without DTLS    
    result = pycoap.Request("localhost:5683", "/.well-known/core")
    print(result)

def sendPayload():
    print(pycoap.DTLSPutRequest("192.168.1.15:5684", "15001/65540", "44d68a62-e6d5-4743-a4fb-ba1317c0e7a5", "eQrKSrpoWgdOPIbw", "{ \"3311\": [{ \"5850\": 1 }] }"))

if __name__ == '__main__':
    result = None
    args = parser.parse_args()

    uri = args.uri.split("/")

    try:
        if not (uri[0] == "coap:" or uri[0] == "coaps:"):
            raise (MalformedURI())
        if not uri[1]=="":
            raise (MalformedURI("Missing //"))
        if not uri[2].find(":")>0:
            raise (MalformedURI("Missing port"))

        dest = "/".join(uri[3:])
        
        if uri[0]=="coap:":
            result = pycoap.Request(uri[2], dest)

        if uri[0]=="coaps:":
            if args.ident==None or args.key==None:
                raise MissingCredentials
            
            result = pycoap.DTLSRequest(uri[2], dest, args.ident, args.key)

        print(result)
    


    except MissingCredentials:
        print("Error: Missing credentials for DTLS-connection!")
    except MalformedURI:
        print("Error: Malformed uri {0}".format(args.uri))
