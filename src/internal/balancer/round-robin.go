package balancer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
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

func (b *RoundRobinBalancer) HandleRequest(w http.ResponseWriter, r *http.Request) {
	execTime := r.Header.Get("Execution-Time")
	if execTime == "" {
		execTime = "0"
	}

	server, err := b.GetNextServer()
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not get server: %v", err), http.StatusServiceUnavailable)
		return
	}

	if _, err := server.CheckHealth(); err != nil {
		log.Printf("Server %d unavailable: %v", server.ID, err)
		http.Error(w, fmt.Sprintf("Server %d unavailable", server.ID), http.StatusBadGateway)
		server, err = b.GetNextServer()
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not get server: %v", err), http.StatusServiceUnavailable)
			return
		}
		return
	}

	log.Printf("Routing request to server %d, task time: %s", server.ID, execTime)

	// Сразу отвечаем клиенту, что запрос принят
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Request accepted and being processed by server %d\n", server.ID)

	// Создаем новый контекст без привязки к клиентскому запросу
	ctx := context.Background()

	// Клонируем запрос с новым контекстом
	req := r.Clone(ctx)

	// Создаем прокси
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Host = server.URL[len("http://"):]
			r.URL.Path = "/process"
			r.Header.Set("Execution-Time", execTime)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error (server %d): %v", server.ID, err)
		},
	}

	// Запускаем обработку в горутине
	go func() {
		// Используем httptest.NewRecorder() как заглушку для ResponseWriter
		proxy.ServeHTTP(httptest.NewRecorder(), req)
		// log.Printf("Request processed by server %d", server.ID)
	}()
}

// func (b *RoundRobinBalancer) HandleRequest(w http.ResponseWriter, r *http.Request) {
// 	// Получаем время выполнения из заголовка
// 	execTime := r.Header.Get("Execution-Time")
// 	if execTime == "" {
// 		execTime = "0"
// 	}

// 	// Выбираем сервер
// 	server, err := b.GetNextServer()
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Could not get server: %v", err), http.StatusServiceUnavailable)
// 		return
// 	}

// 	// Проверяем здоровье сервера
// 	if _, err := server.CheckHealth(); err != nil {
// 		log.Printf("Server %d unavailable: %v", server.ID, err)
// 		http.Error(w, fmt.Sprintf("Server %d unavailable", server.ID), http.StatusBadGateway)
// 		return
// 	}

// 	log.Printf("Routing request to server %d, task time: %s", server.ID, execTime)

// 	// Создаем прокси с обработчиком ошибок
// 	proxy := &httputil.ReverseProxy{
// 		Director: func(r *http.Request) {
// 			r.URL.Scheme = "http"
// 			r.URL.Host = server.URL[len("http://"):]
// 			r.URL.Path = "/process"
// 			r.Header.Set("Execution-Time", execTime)
// 		},
// 		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
// 			log.Printf("Proxy error (server %d): %v", server.ID, err)
// 			http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		},
// 	}

// 	// Обрабатываем запрос
// 	proxy.ServeHTTP(w, r)
// }
