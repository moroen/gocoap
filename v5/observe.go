package gocoap

import (
	"context"
	"log"
	"time"

	"github.com/plgd-dev/go-coap/v2/udp/message/pool"
)

func ObserveStop()                  {}
func ObserveRestart(reconnect bool) {}

func Observe(params ObserveParams) error {
	sync := make(chan bool)

	co, err := getDTLSConnection(RequestParams{Host: params.Host, Port: params.Port, Id: params.Id, Key: params.Key})
	if err != nil {
		return err
	}

	num := 0
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	obs, err := co.Observe(ctx, "/15001/65542", func(req *pool.Message) {
		log.Printf("Got %+v\n", req)
		num++
		m, err := req.ReadBody()
		if err != nil {
			log.Fatal(err.Error())
		}

		log.Printf("%s", m)
		/*
			if num >= 10 {
				sync <- true
			}
		*/
	})

	if err != nil {
		log.Fatalf("Unexpected error '%v'", err)
	}
	<-sync
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	obs.Cancel(ctx)

	return nil
}
