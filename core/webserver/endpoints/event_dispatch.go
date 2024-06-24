package endpoints

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/FMotalleb/crontab-go/core/global"
)

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
	event := c.Param("event")
	listeners := global.CTX.EventListeners()[event]
	if len(listeners) == 0 {
		c.String(http.StatusNotFound, fmt.Sprintf("event: '%s' not found", event))
		return
	}
	for _, listener := range listeners {
		go listener()
	}
	c.String(200, fmt.Sprintf("event: '%s' emitted", event))
}
