package api

import (
	"event-data-pipeline/pkg/cli"
	"event-data-pipeline/pkg/logger"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Service struct {
	port     int
	address  string
	scheme   string
	basePath string
	router   *gin.Engine
}

func NewService() *Service {

	mode := "release"
	if cli.Args.DebugEnabled {
		mode = "debug"
	}

	gin.SetMode(mode)
	// Create a gin router
	router := gin.New()

	svc := &Service{
		port:     cli.Args.Port,
		address:  cli.Args.Addr,
		scheme:   cli.Args.Scheme,
		basePath: cli.Args.BasePath,
		router:   router,
	}

	NewRouteHandler(svc)

	return svc
}

func (s *Service) Run() {
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", s.port),
		Handler:        s.router,
		MaxHeaderBytes: 1 << 20,
	}
	logger.Infof("Number of routes: %d", len(s.router.Routes()))
	for _, r := range s.router.Routes() {
		logger.Infof("route: %s %s", r.Method, r.Path)
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		logger.Infof("Listening on: %s", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			logger.Fatalf("Failed to start service: %v", err)
		}
	}()

}
