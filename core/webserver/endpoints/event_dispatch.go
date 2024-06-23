package endpoints

import "github.com/gin-gonic/gin"

type EventChannel = chan string

type EventDispatchEndpoint struct {
	eventChan EventChannel
}

func NewEventDispatchEndpoint() *EventDispatchEndpoint {
	return &EventDispatchEndpoint{
		eventChan: make(EventChannel),
	}
}

func (ed *EventDispatchEndpoint) Endpoint(c *gin.Context) {
	c.Request.PathValue("event")
}
