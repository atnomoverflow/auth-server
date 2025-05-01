package v1

import (
	"net/http"

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

func (h *Handler) Init(api *http.ServeMux) {
	// init other routes
	v1 := http.NewServeMux()
	api.Handle("/v1/", http.StripPrefix("/v1", v1))

	h.initUserRoutes(v1)
	h.initAuthRoutes(v1)

}
