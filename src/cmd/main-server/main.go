package main

import (
	"fmt"
	"net/http"

	"github.com/pozedorum/load_balancer/internal/balancer"
	"github.com/pozedorum/load_balancer/pkg/server"
)

func main() {
	servers := server.CreateServers(5) // Создание 5 серверов
	balancer := balancer.NewRoundRobinBalancer(servers)

	http.HandleFunc("/", balancer.HandleRequest)
	fmt.Println("Load Balancer started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
