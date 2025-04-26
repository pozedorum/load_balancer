package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// server - структура для имитации сервера
type Server struct {
	ID      int
	URL     string       // Адрес сервера (например, "http://localhost:8081")
	Client  *http.Client // HTTP-клиент для health check
	mu      sync.RWMutex // Мьютекс для защиты данных
	Healthy bool         // Флаг здоровья
}

func New(port string) *Server {
	id, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(fmt.Printf("Error with loading server on port %s, error: %w", port, err))
	}
	id -= 8080
	return &Server{
		ID:      id,
		Healthy: true,
		URL:     fmt.Sprintf("http://localhost:%s", port), // Пример: 8081, 8082, ...
		Client:  &http.Client{Timeout: 2 * time.Second},
	}
}

// cодание списка серверов
func CreateServers(count int) []*Server {
	servers := make([]*Server, 0, count)
	for i := 1; i <= count; i++ {
		servers = append(servers, New(strconv.Itoa(i+8080)))
	}
	return servers
}

func (s *Server) CheckHealth() (int, error) {
	resp, err := s.Client.Get(s.URL + "/health")
	if err != nil {
		s.setHealthy(false)
		log.Printf("Server %d health check failed: %v", s.ID, err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.setHealthy(false)
		log.Printf("Server %d unhealthy status: %d", s.ID, resp.StatusCode)
		return resp.StatusCode, fmt.Errorf("status %d", resp.StatusCode)
	}

	s.setHealthy(true)
	return resp.StatusCode, nil
}

func (s *Server) setHealthy(health bool) {
	s.mu.Lock() // полная блокировка
	s.Healthy = health
	s.mu.Unlock()
}

func (s *Server) IsHealthy() bool {
	s.mu.RLock() // блокировка чтения
	defer s.mu.RUnlock()
	return s.Healthy
}

// Новая функция для обработки задачи
func (s *Server) ProcessTask(delay time.Duration) time.Duration {
	// Имитация обработки с случайной задержкой
	time.Sleep(delay)
	return delay
}
