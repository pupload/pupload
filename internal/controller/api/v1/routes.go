package v1

import (
	"net/http"

	flows "github.com/pupload/pupload/internal/controller/flows/service"

	"github.com/go-chi/chi/v5"
)

func HandleAPIRoutes(f *flows.FlowService) http.Handler {

	r := chi.NewRouter()

	r.Mount("/flow", handleFlowRoutes(f))
	r.Mount("/upload", handleUploadRoutes())

	return r

}
