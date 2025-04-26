package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/pozedorum/load_balancer/internal/balancer"
	"github.com/pozedorum/load_balancer/internal/server"
)

func main() {
	// Создаем 3 сервера через server.New()
	servers := make([]*server.Server, 0, 3)
	for port := 8081; port <= 8083; port++ {
		servers = append(servers, server.New(strconv.Itoa(port)))
	}

	// Инициализируем балансировщик
	lb := balancer.NewRoundRobinBalancer(servers)

	// Настраиваем HTTP-сервер балансировщика
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := lb.HandleRequest(w, r); err != nil {
			log.Printf("Error handling request: %v", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "Error: %v\n", err)
		}
	})

	port := 8080
	log.Printf("Load balancer started on :%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
