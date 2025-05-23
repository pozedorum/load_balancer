package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/pozedorum/load_balancer/pkg/logger"
)

const logDir = "logs"

// server - структура сервера
type Server struct {
	ID      int          // Номер сервера в списке серверов балансировщика
	URL     string       // Адрес сервера (например, "http://localhost:8081")
	Client  *http.Client // HTTP-клиент для health check
	Logger  *log.Logger  // Логгер
	mu      sync.RWMutex // Мьютекс для защиты данных
	Healthy bool         // Флаг здоровья
}

// Конструктор сервера со стороны балансировщика
func New(port string) *Server {
	id, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Error with loading server on port %s, error: %w", port, err)
	}
	id -= 8080
	return &Server{
		ID:      id,
		Healthy: true,
		URL:     fmt.Sprintf("http://localhost:%s", port), // Пример: 8081, 8082, ...
		Client:  &http.Client{Timeout: 2 * time.Second},
	}
}

// Конструктор сервера со стороны сервера
func NewWithLogger(port string, logger *logger.Logger) *Server {
	id, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Error with loading server on port %s, error: %w", port, err)
	}
	id -= 8080
	return &Server{
		ID:      id,
		Healthy: true,
		URL:     fmt.Sprintf("http://localhost:%s", port), // Пример: 8081, 8082, ...
		Client:  &http.Client{Timeout: 2 * time.Second},
		Logger:  logger.Logger,
	}
}

// cоздание списка серверов
func CreateServers(count int) []*Server {
	servers := make([]*Server, 0, count)
	for i := 1; i <= count; i++ {
		servers = append(servers, New(strconv.Itoa(i+8080)))
	}
	return servers
}

// отправка запроса на проверку состояния сервера и обработка ответа
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

// // функция установки статуса с блокировкой данных
func (s *Server) setHealthy(health bool) {
	s.mu.Lock() // полная блокировка
	s.Healthy = health
	s.mu.Unlock()
}

// функция проверки состояния здоровья и блокировки только на чтение данных
func (s *Server) IsHealthy() bool {
	s.mu.RLock() // блокировка чтения
	defer s.mu.RUnlock()
	return s.Healthy
}

// Новая функция для обработки задачи
func (s *Server) ProcessTask(delay time.Duration) time.Duration {
	// Имитация обработки с случайной задержкой
	time.Sleep(delay * time.Millisecond)
	return delay
}
