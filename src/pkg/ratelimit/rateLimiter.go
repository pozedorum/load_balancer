package ratelimit

import (
	"log"
	"sync"
	"time"

	"github.com/pozedorum/load_balancer/config"
)

const RateLimitsConfigPath = "config/rate_limits.json"

// RateLimiter представляет собой модуль rate-limiting
type RateLimiter struct {
	mu              sync.RWMutex       // мьютекс защиты данных
	clients         map[string]*Client // список клиентов (в клиентах лежат их бакеты)
	cleanupInterval time.Duration      // Интервал очистки
	inactiveTimeout time.Duration      // Таймаут неактивности
	stopChan        chan struct{}      // Канал для остановки очистки
	defaultCapacity int
	defaultRate     int
}

// NewRateLimiter создает новый модуль rate-limiting
func NewRateLimiter(cleanupInterval, inactiveTimeout time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients:         make(map[string]*Client),
		cleanupInterval: cleanupInterval,
		inactiveTimeout: inactiveTimeout,
		stopChan:        make(chan struct{}),
		defaultCapacity: 10, // значения по умолчанию
		defaultRate:     1,
	}
	go rl.startCleanup()
	return rl
}

func NewRateLimiterWithConfig(cleanupInterval, inactiveTimeout time.Duration) *RateLimiter {
	rl := NewRateLimiter(cleanupInterval, inactiveTimeout)

	// Загружаем и применяем конфиг
	if cfg, err := config.LoadRateLimitConfig(RateLimitsConfigPath); err == nil {
		rl.defaultCapacity = cfg.Default.Capacity
		rl.defaultRate = cfg.Default.Rate

		// Предварительно создаем клиентов из конфига
		for ip, clientCfg := range cfg.Clients {
			rl.clients[ip] = NewClient(ip, clientCfg.Capacity, clientCfg.Rate)
		}
	} else {
		// Логируем ошибку, если конфиг не загрузился
		log.Printf("Failed to load rate limit config: %v. Using defaults", err)
	}

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
		// Если клиента нет, создаем с дефолтными настройками
		client = NewClient(ip, rl.defaultCapacity, rl.defaultRate)
		rl.clients[ip] = client
	}
	// в клиенте свой мьютекс, так что здесь его надо разблокировать
	rl.mu.Unlock()

	client.UpdateLastSeen()
	return client.TakeToken()
}

// Возват токена в случае невыполнения запроса (возникла ошибка при выполнении)
func (r *RateLimiter) ReturnToken(ip string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if client, exists := r.clients[ip]; exists {
		client.ReturnToken()
	}
}

// Разрешение на взятие токена из бакета (разрешение на выполнение запроса)
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
