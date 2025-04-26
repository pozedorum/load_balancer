package balancer

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/pozedorum/load_balancer/internal/server"
)

var (
	ErrNoHealthyServers = errors.New("no healthy servers available")
	ErrInvalidResponse  = errors.New("invalid server response")
)

type RoundRobinBalancer struct {
	servers []*server.Server
	lock    sync.Mutex
	current int
}

func NewRoundRobinBalancer(servers []*server.Server) *RoundRobinBalancer {
	balancer := &RoundRobinBalancer{servers: servers}
	go balancer.StartHealthCheck()
	return balancer
}

func (b *RoundRobinBalancer) StartHealthCheck() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, s := range b.servers {
			go s.CheckHealth()
		}
	}
}

func (b *RoundRobinBalancer) GetNextServer() (*server.Server, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	for i := 0; i < len(b.servers); i++ {
		server := b.servers[b.current]

		if server.Healthy {
			b.current = (b.current + 1) % len(b.servers)
			return server, nil
		}
	}

	return nil, ErrNoHealthyServers
}

// обработка запроса балансировщиком
func (b *RoundRobinBalancer) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	execTime := r.Header.Get("Execution-Time")
	if execTime == "" {
		execTime = "0"
	}

	server, err := b.GetNextServer()
	if err != nil {
		return fmt.Errorf("could not get server: %w", err)
	}

	code, err := server.CheckHealth()
	if err != nil {
		log.Printf("Server %d health check failed (code %d): %v", server.ID, code, err)
		return fmt.Errorf("server unavailable")
	}

	log.Printf("Routing request to server %d, task time: %s", server.ID, execTime)

	// Создаем обратный прокси-сервер
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Host = server.URL[len("http://"):] // Удалить схему и порт
			r.Header.Set("Execution-Time", execTime)
		},
	}

	proxy.ServeHTTP(w, r)
	return nil
}
