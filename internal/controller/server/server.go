package controllerserver

import (
	"net/http"
	v1 "pupload/internal/controller/api/v1"
	config "pupload/internal/controller/config"
	flows "pupload/internal/controller/flows"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewServer(config config.CONTROLLER_CONFIG) http.Handler {
	r := chi.NewRouter()

	flows.CreateFlowService(config.Storage.DataPath)

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Mount("/api/v1", v1.HandleAPIRoutes())
	return r
}
