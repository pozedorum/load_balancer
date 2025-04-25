package balancer

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/pozedorum/load_balancer/pkg/server"
)

type RoundRobinBalancer struct {
	servers []*server.Server
	current int
	lock    sync.Mutex
}

func NewRoundRobinBalancer(servers []*server.Server) *RoundRobinBalancer {
	return &RoundRobinBalancer{servers: servers}
}

func (b *RoundRobinBalancer) GetNextServer() *server.Server {
	b.lock.Lock()
	defer b.lock.Unlock()

	server := b.servers[b.current]
	b.current = (b.current + 1) % len(b.servers)
	return server
}

// обработка запроса балансировщиком
func (b *RoundRobinBalancer) HandleRequest(w http.ResponseWriter, r *http.Request) {
	server := b.GetNextServer()
	server.HandleRequest(w, r)
}

func main() {
	servers := server.CreateServers(5) // Создание 5 серверов
	balancer := NewRoundRobinBalancer(servers)

	http.HandleFunc("/", balancer.HandleRequest)
	fmt.Println("Load Balancer started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
