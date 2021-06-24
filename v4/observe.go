package gocoap

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/plgd-dev/go-coap/v2/udp/client"
	"github.com/plgd-dev/go-coap/v2/udp/message/pool"
)

var _observe_connection *client.ClientConn

var observe_params ObserveParams
var observe_callback func([]byte) error
var sync chan (bool)

func Observe(params ObserveParams, callback func([]byte) error) error {
	observe_params = params
	observe_callback = callback
	ObserveStart()
	return nil
}

func ObserveStart() error {
	doObserve(observe_params, observe_callback)
	return nil
}

func ObserveStop() error {
	if sync != nil {
		close(sync)
	}
	return nil
}

func ObserveRestart(reconnect bool) error {

	log.WithFields(log.Fields{
		"reconnect": reconnect,
	}).Debug("Observe restart called")

	ObserveStop()
	if reconnect {
		closeDTLSObserveConnection()
	}
	ObserveStart()
	return nil
}

func getDTLSObserveConnection(param RequestParams) (*client.ClientConn, error) {
	if _observe_connection != nil {
		// log.Println("Using old connection")
		return _connection, nil
	}
	log.Debug("getDTLSObserveConnection: Creating new connection")
	_observe_connection, err := createDTLSConnection(param)
	if err != nil {
		err = ErrorHandshake
	}

	return _observe_connection, err
}

func closeDTLSObserveConnection() error {
	if _observe_connection != nil {
		log.Debug("Observe-connection closing")
		err := _connection.Close()
		if err != nil {
			return err
		}
		_observe_connection = nil
	}
	return nil
}

func doObserve(params ObserveParams, callback func(b []byte) error) error {

	defer closeDTLSObserveConnection()

	co, err := getDTLSObserveConnection(RequestParams{Host: params.Host, Port: params.Port, Id: params.Id, Key: params.Key})
	if err != nil {
		return err
	}

	sync = make(chan bool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for _, uri := range params.Uri {

		go func(uri string, stop chan bool) {
			log.WithFields(log.Fields{
				"endpoint": uri,
			}).Debug("Starting observe")

			obs, err := co.Observe(ctx, uri, func(req *pool.Message) {
				m, err := req.ReadBody()
				if err != nil {
					log.Fatal(err.Error())
				}

				observe_callback(m)

			})

			if err != nil {
				log.Fatalf("Unexpected error '%v'", err)
			}
			<-stop
			log.WithFields(log.Fields{
				"endpoint": uri,
			}).Debug("Stopping observe")

			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			obs.Cancel(ctx)
		}(uri, sync)
	}
	<-sync
	log.Println("Observe sync done")
	return nil
}
