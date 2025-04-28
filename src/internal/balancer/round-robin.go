package balancer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"sync"
	"time"

	"github.com/pozedorum/load_balancer/internal/server"
	"github.com/pozedorum/load_balancer/pkg/ratelimit"
)

var (
	ErrNoHealthyServers = errors.New("no healthy servers available")
	ErrInvalidResponse  = errors.New("invalid server response")
)

// структура балансировщика
type RoundRobinBalancer struct {
	servers     []*server.Server       // список серверов
	rateLimiter *ratelimit.RateLimiter // ограничитель количества запросов
	lock        sync.Mutex             // мьютекс блокировки данных
	current     int                    // текущий сервер выбранный для отправки запроса
}

// конструктор балансировщика
func NewRoundRobinBalancer(servers []*server.Server) *RoundRobinBalancer {
	balancer := &RoundRobinBalancer{
		servers:     servers,
		rateLimiter: ratelimit.NewRateLimiter(5*time.Minute, 5*time.Minute),
	}
	go balancer.StartHealthCheck()
	return balancer
}

// функция автоматической проверки состояния серверов
func (b *RoundRobinBalancer) StartHealthCheck() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, s := range b.servers {
			go s.CheckHealth()
		}
	}
}

// функция получения сервера из списка серверов
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
	clientIP := strings.Split(r.RemoteAddr, ":")[0]
	if !b.rateLimiter.TakeToken(clientIP) {
		log.Printf("request from %s is canceled", clientIP)
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}
	var err error
	defer func() {
		if err != nil {
			b.rateLimiter.ReturnToken(clientIP)
		}
	}()

	execTime := r.Header.Get("Execution-Time")
	if execTime == "" {
		execTime = "0"
	}

	// Поиск здорового сервера
	var server *server.Server

	for range len(b.servers) - 1 {
		server, err = b.GetNextServer()
		if err != nil {
			continue
		}
		if _, err = server.CheckHealth(); err == nil {
			break
		}
	}
	if err != nil {
		http.Error(w, "No healthy servers available", http.StatusServiceUnavailable)
		return
	}

	log.Printf("Routing request to server %d, task time: %s", server.ID, execTime)

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Request accepted and being processed by server %d\n", server.ID)

	ctx := context.Background()
	req := r.Clone(ctx)

	// Копирование важных заголовков
	for _, h := range []string{"Accept", "Accept-Encoding", "Content-Type"} {
		if v := r.Header.Get(h); v != "" {
			req.Header.Set(h, v)
		}
	}

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

	// Буферизированный обработчик
	recorder := httptest.NewRecorder()
	go func() {
		proxy.ServeHTTP(recorder, req)
		if recorder.Code >= 400 {
			log.Printf("Backend %d response: %d - %s",
				server.ID, recorder.Code, recorder.Body.String())
		}
	}()
}
