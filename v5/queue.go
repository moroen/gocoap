package gocoap

import (
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func (c *CoapDTLSConnection) AddToQueue(request CoapDTLSRequest) {
	c.mu.Lock()
	c.queue = append(c.queue, request)
	c.mu.Unlock()
}

func (c *CoapDTLSConnection) QueueLenght() int {
	c.mu.Lock()

	defer c.mu.Unlock()
	return len(c.queue)
}

func (c *CoapDTLSConnection) HandleQueue() {
	defer c.mu.Unlock()

	log.WithFields(log.Fields{
		"Items": len(c.queue),
	}).Debug("Tradfri: HandleQueue")

	var item CoapDTLSRequest
	c.mu.Lock()
	for len(c.queue) > 0 {
		item, c.queue = c.queue[0], c.queue[1:]
		switch item.RequestMethod {
		case "GET":
			ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
			c.GET(ctx, item.Uri, item.Handler)
		}
	}
}
