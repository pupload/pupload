package v1

import (
	"fmt"
	"net/http"
	flows "pupload/internal/controller/flows/service"
	"pupload/internal/logging"
	"pupload/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func handleFlowRoutes(f *flows.FlowService) http.Handler {

	log := logging.ForService("api")

	r := chi.NewRouter()

	// r.Get("/list", func(w http.ResponseWriter, r *http.Request) {
	// 	render.JSON(w, r, flows)
	// })

	// r.Post("/execute/{flowName}", func(w http.ResponseWriter, r *http.Request) {
	// 	flowName := chi.URLParam(r, "flowName")
	// 	var input map[string]interface{}
	// 	if err := render.DecodeJSON(r.Body, &input); err != nil {
	// 		http.Error(w, "Invalid input", http.StatusBadRequest)
	// 		return
	// 	}

	// 	taskID, err := f.StartFlow(r.Context(), flowName)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}

	// 	render.JSON(w, r, map[string]string{"task_id": taskID})
	// })

	r.Get("/status/{flowRunID}", func(w http.ResponseWriter, r *http.Request) {
		flowRunID := chi.URLParam(r, "flowRunID")
		flowRunStatus, err := f.Status(flowRunID)
		if err != nil {
			http.Error(w, "Flow run does not exist.", http.StatusBadRequest)
		}

		render.JSON(w, r, flowRunStatus)

	})

	r.Post("/test", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Flow     models.Flow
			NodeDefs []models.NodeDef
		}

		if err := render.DecodeJSON(r.Body, &input); err != nil {
			log.Error("error unwrapping json", "err", err)
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		run, err := f.RunFlow(input.Flow, input.NodeDefs)
		if err != nil {
			log.Error("unable to run flow", "err", err)
			http.Error(w, fmt.Sprintf("unable to run flow: %s", err), http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, run)

	})

	return r

}
