package gocoap

import "errors"

// ErrorTimeout error
var ErrorTimeout = errors.New("COAP Error: Connection timeout")

// ErrorBadIdent error
var ErrorBadIdent = errors.New("COAP DTLS Error: Wrong credentials?")

// ErrorHandshake error
var ErrorHandshake = errors.New("COAP DTLS Error: Handshake timeout")

// ErrorReadTimeout error
var ErrorReadTimeout = errors.New("COAP DTLS Error: Read timeout")

// ErrorWriteTimeout error
var ErrorWriteTimeout = errors.New("COAP DTLS Error: Write timeout")

// ErrorNoConfig error
var ErrorNoConfig = errors.New("COAP Error: No config")

// MethodNotAllowed error
var MethodNotAllowed = errors.New("COAP Error: Method not allowed")

// UriNotFound
var UriNotFound = errors.New("COAP Error: Uri not found")

// Unauthorized
var Unauthorized = errors.New("COAP Error: Unauthorized")

// BadRequest
var BadRequest = errors.New("COAP Error: Bad Request")

// ErrorBadData
var ErrorBadData = errors.New("COAP Error: Bad returndata")

// ErrorUnknowError
var ErrorUnknownError = errors.New("COAP Error: Unknown status")

// ErrorNoPayload
var ErrorNoPayload = errors.New("COAP Error: No payload for request")
