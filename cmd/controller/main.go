package main

import (
	"net/http"
	config "pupload/internal/controller/config"
	controllerserver "pupload/internal/controller/server"
)

func main() {

	config := config.LoadControllerConfig("/home/seb/Projects/OpenUpload/test_config")

	srv := &http.Server{
		Addr:    ":1234",
		Handler: controllerserver.NewServer(config),
	}

	srv.ListenAndServe()

}
