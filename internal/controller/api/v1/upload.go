package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func handleUploadRoutes() http.Handler {

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {

	})

	return r
}
