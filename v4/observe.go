package gocoap

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/plgd-dev/go-coap/v2/udp/client"
	"github.com/plgd-dev/go-coap/v2/udp/message/pool"
)

// var _observe_connection *client.ClientConn
var _wgObserve sync.WaitGroup

var observe_params ObserveParams
var observe_callback func([]byte) error

var control context.Context
var status context.Context
var observe_stop func()
var stop_done func()

func Observe(params ObserveParams, callback func([]byte) error) error {
	observe_params = params
	observe_callback = callback

	fmt.Println(params.KeepAlive)

	control, observe_stop = context.WithCancel(context.Background())
	status, stop_done = context.WithCancel(context.Background())

	doObserve(params, callback)

	return nil
}

func ObserveStop() error {
	if observe_stop == nil {
		// Observe goroutine not started
		return errors.New("Observe goroutine not started")
	}
	observe_stop()
	<-status.Done()
	fmt.Println("Stop done")
	return nil
}

func ObserveRestart(reconnect bool) error {
	if err := ObserveStop(); err == nil {
		Observe(observe_params, observe_callback)
		return nil
	} else {
		return err
	}
}

/*
func getDTLSObserveConnection(param RequestParams) (*client.ClientConn, error) {
	var err error

	if _observe_connection != nil {
		log.Println("Using old connection")
		return _observe_connection, nil
	}

	log.Debug("getDTLSObserveConnection: Creating new connection")

	retry := 1

	for retry > -1 {
		if _observe_connection, err = createDTLSConnection(param); err == nil {
			log.Info(fmt.Sprintf("Connected to tradfri for observe at [tcp://%s:%d]", param.Host, param.Port))
			return _observe_connection, nil
		} else {
			_observe_connection = nil
			log.WithFields(log.Fields{
				"error": err.Error(),
				"try":   retry,
			}).Error("getDTLSObserveConnection")
			retry++
			time.Sleep(5 * time.Second)
		}
	}
	return nil, ErrorHandshake
}

func closeDTLSObserveConnection() error {
	if _observe_connection != nil {
		log.Debug("Closing observe-connection")
		err := _observe_connection.Close()
		if err != nil {
			return err
		}
		_observe_connection = nil
	}
	return nil
}
*/

func doObserve(params ObserveParams, callback func(b []byte) error) error {

	co, err := getDTLSConnection(RequestParams{Host: params.Host, Port: params.Port, Id: params.Id, Key: params.Key})
	if err != nil {
		return err
	}

	for _, uri := range params.Uri {
		_wgObserve.Add(1)

		go func(uri string, keepAlive int) {
			var obs *client.Observation
			var ticker time.Ticker

			if params.KeepAlive > 0 {
				ticker = *time.NewTicker(time.Duration(params.KeepAlive) * time.Second)
			}

			defer _wgObserve.Done()

		loop:
			for {
				log.WithFields(log.Fields{
					"endpoint": uri,
				}).Debug("Starting observe")

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				obs, err = co.Observe(ctx, uri, func(req *pool.Message) {
					m, err := req.ReadBody()
					if err != nil {
						log.Fatal(err.Error())
					}

					observe_callback(m)
				})

				if err != nil {
					log.WithFields(log.Fields{
						"Error": err.Error(),
						"uri":   uri,
					}).Error("Starting observe")
					observe_stop()
				}

				select {
				case <-ticker.C:
					fmt.Println("Tick")
					obs.Cancel(ctx)
					log.WithFields(log.Fields{
						"endpoint": uri,
					}).Debug("Observe stopped")

				case <-control.Done():
					if obs != nil {
						obs.Cancel(ctx)
						log.WithFields(log.Fields{
							"endpoint": uri,
						}).Debug("Observe stopped")
					} else {
						log.WithFields(log.Fields{
							"endpoint": uri,
						}).Debug("Stopping observe failed")
					}
					break loop
				}
			}
		}(uri, params.KeepAlive)
	}
	_wgObserve.Wait()
	stop_done()
	return nil
}
