package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/x0ddf/whodidthat-controller/controllers"
)

func main() {
	var port int
	var tlsKey, tlsCert string
	flag.IntVar(&port, "port", 8443, "The port to listen on")
	flag.StringVar(&tlsKey, "tls-key", "/etc/webhook/certs/tls.key", "Private key for TLS")
	flag.StringVar(&tlsCert, "tls-crt", "/etc/webhook/certs/tls.crt", "TLS certificate")
	flag.Parse()
	certs, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	if err != nil {
		log.Fatalf("fail to load tls certificates:%v", err)
	}
	ac := controllers.NewAdmissionController()

	http.HandleFunc("/mutate", ac.Handle)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
		if err != nil {
			return
		}
	})

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{certs},
		},
	}
	log.Printf("starting server on %d", port)
	if err := server.ListenAndServeTLS(tlsCert, tlsKey); err != nil {
		log.Panic(err)
	}
}
