package server

import (
	"fmt"
	"net/http"
	"time"
)

// srver - структура для имитации сервера
type Server struct {
	ID int
}

// handleRequest - обработка запроса сервером
func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Server %d handled request\n", s.ID)
	time.Sleep(100 * time.Millisecond) // Имитация обработки
}

// cодание списка серверов
func createServers(count int) []*Server {
	var servers []*Server
	for i := 0; i < count; i++ {
		servers = append(servers, &Server{ID: i + 1})
	}
	return servers
}
