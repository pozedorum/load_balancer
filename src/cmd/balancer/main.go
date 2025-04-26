package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/pozedorum/load_balancer/config"
	"github.com/pozedorum/load_balancer/internal/balancer"
	"github.com/pozedorum/load_balancer/internal/server"
)

func main() {
	configs, err := config.LoadServerConfigList("config/servers.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	servers := make([]*server.Server, 0, len(configs))
	for _, cfg := range configs {
		servers = append(servers, server.New(strconv.Itoa(cfg.Port)))
	}

	lb := balancer.NewRoundRobinBalancer(servers)

	// Упрощенный обработчик без возврата ошибки
	http.HandleFunc("/", lb.HandleRequest)

	port := 8080
	log.Printf("Load balancer started on :%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
