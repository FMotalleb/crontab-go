// Package endpoint implements the logic behind each endpoint of the webserver
package endpoint

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/FMotalleb/crontab-go/core/global"
)

type EventDispatchEndpoint struct{}

func NewEventDispatchEndpoint() *EventDispatchEndpoint {
	return &EventDispatchEndpoint{}
}

func (ed *EventDispatchEndpoint) Endpoint(c *gin.Context) {
	event := c.Param("event")
	listeners := global.CTX.EventListeners()[event]
	if len(listeners) == 0 {
		c.String(http.StatusNotFound, fmt.Sprintf("event: '%s' not found", event))
		return
	}
	listenerCount := len(listeners)
	for _, listener := range listeners {
		go listener()
	}
	c.String(http.StatusOK, fmt.Sprintf("event: '%s' emitted, %d listeners where found", event, listenerCount))
}
