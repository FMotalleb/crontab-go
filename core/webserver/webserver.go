// Package webserver implements the logic for the webserver
package webserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/core/webserver/endpoint"
	"github.com/FMotalleb/crontab-go/helpers"
	"github.com/FMotalleb/crontab-go/logger"
)

type AuthConfig struct {
	Username string
	Password string
}

type WebServer struct {
	*AuthConfig
	ctx          context.Context
	address      string
	port         uint
	log          *logrus.Entry
	serveMetrics bool
}

func NewWebServer(ctx context.Context,
	address string,
	port uint,
	serveMetrics bool,
	authentication *AuthConfig,
) *WebServer {
	return &WebServer{
		ctx:          ctx,
		address:      address,
		port:         port,
		AuthConfig:   authentication,
		log:          logger.SetupLogger("WebServer"),
		serveMetrics: serveMetrics,
	}
}

func (s *WebServer) Serve() {
	engine := gin.New()
	auth := func(*gin.Context) {}
	if s.AuthConfig != nil && s.AuthConfig.Username != "" && s.AuthConfig.Password != "" {
		auth = gin.BasicAuth(gin.Accounts{s.AuthConfig.Username: s.AuthConfig.Password})
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
	if s.serveMetrics {
		engine.GET("/metrics", func(ctx *gin.Context) {
			promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
		})
	} else {
		engine.GET("/metrics", func(ctx *gin.Context) {
			ctx.String(http.StatusNotFound, "Metrics are disabled, please enable metrics using `WEBSERVER_METRICS=true`")
		})
	}

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
