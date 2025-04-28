package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pozedorum/load_balancer/config"
	"github.com/pozedorum/load_balancer/internal/server"
	"github.com/pozedorum/load_balancer/pkg/logger"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run cmd/backend/main.go <config_file> <server_id>")
	}

	configFile := os.Args[1]
	serverID := os.Args[2]

	port, err := config.FindPortInConfig(configFile, serverID)
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация логгера
	logger, err := logger.New("backend", serverID, port)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()
	logger.SetGlobal()

	// Обработка сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Close()
		os.Exit(0)
	}()

	// Создаем сервер с передачей логгера
	srv := server.NewWithLogger(port, logger)
	logger.Printf("Starting server on %s", srv.URL)

	http.HandleFunc("/", srv.HandleRequest)
	logger.Printf("Server is ready to accept connections on :%s", port)
	logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
