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
	ctx      context.Context
	address  string
	port     uint
	username string
	password string
	log      *logrus.Entry
}

func NewWebServer(ctx context.Context, address string, port uint, username string, password string) *WebServer {
	return &WebServer{
		ctx:      ctx,
		address:  address,
		port:     port,
		username: username,
		password: password,
		log:      logger.SetupLogger("WebServer"),
	}
}

func (s *WebServer) Serve() {
	engine := gin.New()
	auth := func(*gin.Context) {}
	if s.username != "" && s.password != "" {
		auth = gin.BasicAuth(gin.Accounts{s.username: s.password})
	} else {
		s.log.Warnf("received no value on username or password, ignoring any authentication, if you intended to use no authentication ignore this message")
	}
	log := gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: s.formatter,
	})
	engine.Use(
		auth,
		log,
		gin.Recovery(),
	)

	engine.GET(
		"/foo",
		func(c *gin.Context) {
			c.String(200, "bar")
		},
	)

	ed := &endpoint.EventDispatchEndpoint{}
	engine.Any(
		"/events/:event/emit",
		ed.Endpoint,
	)

	err := engine.Run(fmt.Sprintf("%s:%d", s.address, s.port))
	helpers.FatalOnErr(s.log, err, "Failed to start webserver: %s")
}

func (s *WebServer) formatter(params gin.LogFormatterParams) string {
	log := s.log.WithFields(
		logrus.Fields{
			"status_code": params.StatusCode,
			"client_ip":   params.ClientIP,
			"method":      params.Method,
			"path":        params.Path,
		},
	)
	if params.ErrorMessage != "" {
		log = s.log.WithFields(
			logrus.Fields{
				"error": params.ErrorMessage,
			},
		)
	}
	log.Level = logrus.DebugLevel
	log.Message = fmt.Sprintf("served a %s request in path: %s", params.Method, params.Path)
	answer, err := log.String()
	if err != nil {
		log.WithError(err).Warn("cannot send log message to gin logger")
		return ""
	}
	return answer
}
