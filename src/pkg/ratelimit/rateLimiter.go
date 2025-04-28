package ratelimit

import (
	"sync"
	"time"
)

// RateLimiter представляет собой модуль rate-limiting
type RateLimiter struct {
	clients         map[string]*Client
	mu              sync.RWMutex
	cleanupInterval time.Duration // Интервал очистки
	inactiveTimeout time.Duration // Таймаут неактивности
	stopChan        chan struct{} // Канал для остановки очистки
}

// NewRateLimiter создает новый модуль rate-limiting
func NewRateLimiter(cleanupInterval, inactiveTimeout time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients:         make(map[string]*Client),
		cleanupInterval: cleanupInterval,
		inactiveTimeout: inactiveTimeout,
		stopChan:        make(chan struct{}),
	}
	go rl.startCleanup()
	return rl
}

// AddClient добавляет нового клиента в модуль rate-limiting
func (r *RateLimiter) AddClient(ip string, capacity, rate int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[ip] = NewClient(ip, capacity, rate)
}

// TakeToken извлекает токен из bucket клиента, если он доступен
func (rl *RateLimiter) TakeToken(ip string) bool {
	rl.mu.Lock()
	client, exists := rl.clients[ip]
	if !exists {
		client = NewClient(ip, 10, 1) // Пример: 10 запросов в секунду
		rl.clients[ip] = client
	}
	rl.mu.Unlock()

	client.UpdateLastSeen()
	return client.TakeToken()
}

func (r *RateLimiter) ReturnToken(ip string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if client, exists := r.clients[ip]; exists {
		client.ReturnToken()
	}
}

func (r *RateLimiter) Allow(ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clients[ip]; !exists {
		r.clients[ip] = NewClient(ip, 10, 1) // Например, 10 запросов в секунду
	}

	return r.clients[ip].TakeToken()
}

// Запуск периодической очистки
func (rl *RateLimiter) startCleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupInactiveClients()
		case <-rl.stopChan:
			return
		}
	}
}

// Остановка очистки
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

// Метод очистки неактивных клиентов
func (rl *RateLimiter) cleanupInactiveClients() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, client := range rl.clients {
		if !client.IsActive(rl.inactiveTimeout) {
			delete(rl.clients, ip)
		}
	}
}
