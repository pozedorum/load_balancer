package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/pozedorum/load_balancer/config"
	"github.com/pozedorum/load_balancer/internal/balancer"
	"github.com/pozedorum/load_balancer/internal/server"
	"github.com/pozedorum/load_balancer/pkg/logger"
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
	// Инициализация логгера
	serverID := "0"
	port := "8080"
	log.Printf("Load balancer started on :%s", port)
	logger, err := logger.New("balancer", serverID, port)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()
	logger.SetGlobal()
	// Упрощенный обработчик без возврата ошибки
	http.HandleFunc("/", lb.HandleRequest)

	log.Printf("Load balancer started on :%s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
