package server

import (
	"event-data-pipeline/pkg/api"
)

type HttpServer struct {
	service *api.Service
}

func NewHttpServer() *HttpServer {
	svc := api.NewService()
	return &HttpServer{svc}
}

func (h *HttpServer) Serve() {
	h.service.Run()
}
