package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func handleFlowRoutes() http.Handler {

	r := chi.NewRouter()

	r.Route("/list", func(r chi.Router) {

	})

	return r

}
