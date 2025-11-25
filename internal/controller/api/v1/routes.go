package v1

import (
	"net/http"
	"pupload/internal/controller/flows"

	"github.com/go-chi/chi/v5"
)

func HandleAPIRoutes(f *flows.FlowService) http.Handler {

	r := chi.NewRouter()

	r.Mount("/flow", handleFlowRoutes(f))
	r.Mount("/upload", handleUploadRoutes())

	return r

}
