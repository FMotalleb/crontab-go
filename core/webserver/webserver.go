package webserver

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/helpers"
	"github.com/FMotalleb/crontab-go/logger"
)

type WebServer struct {
	address string
	port    uint
	token   string
	log     *logrus.Entry
}

func NewWebServer(address string, port uint, token string) *WebServer {
	return &WebServer{
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
	routes.Any(
		"/event/dispatch",
		func(c *gin.Context) {
			// TODO: implement
		},
	)

	err := engine.Run(fmt.Sprintf("%s:%d", s.address, s.port))
	helpers.FatalOnErr(s.log, err, "Failed to start webserver: %s")
}
