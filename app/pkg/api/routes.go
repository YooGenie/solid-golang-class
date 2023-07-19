package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Routes struct {
	router *gin.Engine
}

func NewRouteHandler(s *Service) *Routes {
	routes := &Routes{
		router: s.router,
	}

	//Add health check endpoint
	routes.router.GET(fmt.Sprintf("%s/health", s.basePath), gin.WrapF(func(rw http.ResponseWriter, req *http.Request) {
		process(rw)
	}))

	//Add prometheus metrics endpoint
	routes.router.GET(fmt.Sprintf("%s/metrics", s.basePath), gin.WrapH(promhttp.Handler()))
	return routes
}

// process processes the health check request
func process(rw http.ResponseWriter) {
	i := NewInfo()
	msg, _ := json.Marshal(i)
	fmt.Fprintln(rw, string(msg))
}
