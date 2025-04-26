package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pozedorum/load_balancer/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/backend/main.go <server_id>")
	}

	port := os.Args[1]

	srv := server.New(port)

	log.SetPrefix(fmt.Sprintf("[Backend %d] ", srv.ID))
	log.Printf("Starting on %s", srv.URL)

	http.HandleFunc("/", srv.HandleRequest)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
