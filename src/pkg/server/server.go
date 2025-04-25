package server

import (
	"fmt"
	"net/http"
	"time"
)

// server - структура для имитации сервера
type Server struct {
	ID      int
	Healthy bool         // Флаг здоровья
	URL     string       // Адрес сервера (например, "http://localhost:8081")
	Client  *http.Client // HTTP-клиент для health check
}

// handleRequest - обработка запроса сервером
func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Server %d handled request at %v\n", s.ID, time.Now().Format("15:04:05.000"))
	time.Sleep(2000 * time.Millisecond)
}

func New(id int) *Server {
	return &Server{
		ID:      id + 1,
		Healthy: true,
		URL:     fmt.Sprintf("http://localhost:%d", 8081+id), // Пример: 8081, 8082, ...
		Client:  &http.Client{Timeout: 2 * time.Second},
	}
}

// cодание списка серверов
func CreateServers(count int) []*Server {
	servers := make([]*Server, 0, count)
	for i := 0; i < count; i++ {
		servers = append(servers, New(i))
	}
	return servers
}
