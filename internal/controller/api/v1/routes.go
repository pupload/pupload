package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func HandleAPIRoutes() http.Handler {

	r := chi.NewRouter()

	r.Mount("/flow", handleFlowRoutes())
	r.Mount("/upload", handleUploadRoutes())

	return r

}
