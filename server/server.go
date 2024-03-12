package server

import (
	"log"
	"net/http"
)

func Listen(address string, certFile string, keyFile string) error {
	http.HandleFunc("/webhook", handleWebhook)
	log.Printf("webhook server listening on %s", address)
	return http.ListenAndServeTLS(address, certFile, keyFile, nil)
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	log.Print("webhook received")
}
