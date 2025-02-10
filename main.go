package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"

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

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{certs},
		},
	}

	if err := server.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}
