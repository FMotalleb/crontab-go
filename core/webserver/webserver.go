// Package webserver implements the logic for the webserver
package webserver

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/core/webserver/endpoint"
	"github.com/FMotalleb/crontab-go/helpers"
	"github.com/FMotalleb/crontab-go/logger"
)

type WebServer struct {
	ctx     context.Context
	address string
	port    uint
	token   string
	log     *logrus.Entry
}

func NewWebServer(ctx context.Context, address string, port uint, token string) *WebServer {
	return &WebServer{
		ctx:     ctx,
		address: address,
		port:    port,
		token:   token,
		log:     logger.SetupLogger("WebServer"),
	}
}

func (s *WebServer) Serve() {
	engine := gin.Default()
	auth := gin.BasicAuth(gin.Accounts{"admin": s.token})
	routes := engine.Use(auth)
	routes.GET(
		"/foo",
		func(c *gin.Context) {
			c.String(200, "bar")
		},
	)

	ed := &endpoint.EventDispatchEndpoint{}
	routes.Any(
		"/events/:event/emit",
		ed.Endpoint,
	)

	err := engine.Run(fmt.Sprintf("%s:%d", s.address, s.port))
	helpers.FatalOnErr(s.log, err, "Failed to start webserver: %s")
}
