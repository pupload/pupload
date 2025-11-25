package v1

import (
	"net/http"
	"pupload/internal/controller/flows"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func handleFlowRoutes(f *flows.FlowService) http.Handler {

	r := chi.NewRouter()

	r.Get("/list", func(w http.ResponseWriter, r *http.Request) {
		flows := f.ListFlows()
		render.JSON(w, r, flows)
	})

	r.Post("/execute/{flowName}", func(w http.ResponseWriter, r *http.Request) {
		flowName := chi.URLParam(r, "flowName")
		var input map[string]interface{}
		if err := render.DecodeJSON(r.Body, &input); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		taskID, err := f.StartFlow(r.Context(), flowName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, map[string]string{"task_id": taskID})
	})

	r.Get("/status/{flowRunID}", func(w http.ResponseWriter, r *http.Request) {
		flowRunID := chi.URLParam(r, "flowRunID")
		flowRunStatus, err := f.GetFlowRun(flowRunID)
		if err != nil {
			http.Error(w, "Flow run does not exist.", http.StatusBadRequest)
		}

		render.JSON(w, r, flowRunStatus)

	})

	return r

}
