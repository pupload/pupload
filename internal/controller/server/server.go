package controllerserver

import (
	"log"
	"net/http"
	v1 "pupload/internal/controller/api/v1"
	config "pupload/internal/controller/config"
	flows "pupload/internal/controller/flows"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
)

func NewServer(config config.CONTROLLER_CONFIG) http.Handler {
	r := chi.NewRouter()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	f := flows.CreateFlowService(config.Storage.DataPath, rdb)

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Mount("/api/v1", v1.HandleAPIRoutes(f))

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route)
		return nil
	}

	if err := chi.Walk(r, walkFunc); err != nil {
		log.Printf("Logging err: %s\n", err.Error())
	}

	return r
}
