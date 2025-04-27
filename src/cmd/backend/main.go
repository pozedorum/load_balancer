package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pozedorum/load_balancer/config"
	"github.com/pozedorum/load_balancer/internal/server"
)

var serverLogger *log.Logger

func setupLogger(serverID, port string) *os.File {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	logPath := fmt.Sprintf("%s/backend_%s.log", logDir, port)
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Инициализируем глобальный логгер с нужным префиксом
	serverLogger = log.New(logFile, fmt.Sprintf("[Backend %s] ", serverID), log.LstdFlags|log.Lmicroseconds)

	// Перенаправляем стандартный логгер для legacy-кода
	log.SetOutput(logFile)
	log.SetPrefix(fmt.Sprintf("[Backend %s] ", serverID))

	fmt.Fprintf(logFile, "\n\n=== Server %s started at %s ===\n", serverID, time.Now().Format(time.RFC3339))

	return logFile
}

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

	logFile := setupLogger(serverID, port)
	defer logFile.Close()

	// Обработка сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logFile.Close()
		os.Exit(0)
	}()

	// Создаем сервер с передачей логгера
	srv := server.NewWithLogger(port, serverLogger)
	serverLogger.Printf("Starting server on %s", srv.URL)

	http.HandleFunc("/", srv.HandleRequest)
	serverLogger.Printf("Server is ready to accept connections on :%s", port)
	serverLogger.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
