package handler

import (
	"net/http"

	"github.com/atnomoverflow/auth-server/config"
	v1 "github.com/atnomoverflow/auth-server/internal/handler/v1"
	"github.com/atnomoverflow/auth-server/internal/service"
	"github.com/atnomoverflow/auth-server/pkg/logger"
)

type Handler struct {
	services *service.Services
	logger   logger.ILogger
}

func New(services *service.Services, l logger.ILogger) *Handler {
	return &Handler{
		services: services,
		logger:   l,
	}
}
func (h *Handler) Init(cfg config.Config, router *http.ServeMux) {

	api := http.NewServeMux()
	router.Handle("/api/", http.StripPrefix("/api", api))

	handlerv1 := v1.New(h.services, h.logger)
	handlerv1.Init(api)

}
