package controllerserver

import (
	"net/http"
	"time"

	v1 "github.com/pupload/pupload/internal/controller/api/v1"
	config "github.com/pupload/pupload/internal/controller/config"
	flows "github.com/pupload/pupload/internal/controller/flows/service"
	"github.com/pupload/pupload/internal/logging"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewServer(config config.ControllerSettings, f *flows.FlowService) http.Handler {

	log := logging.ForService("server")

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	// r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Mount("/api/v1", v1.HandleAPIRoutes(f))

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Info("Route", "method", method, "route", route)
		return nil
	}

	chi.Walk(r, walkFunc)

	return r
}
