package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/x0ddf/whodidthat-controller/controllers"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8443, "The port to listen on")
	flag.Parse()

	ac := controllers.NewAdmissionController()

	http.HandleFunc("/mutate", ac.Handle)

	// You'll need to set up TLS certificates for production use
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
