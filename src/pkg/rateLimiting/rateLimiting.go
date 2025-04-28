package ratelimit

import (
	"sync"
	"time"
)

// Bucket представляет собой bucket токенов для клиента
type Bucket struct {
	capacity int          // емкость bucket
	rate     int          // скорость пополнения bucket
	ticker   *time.Ticker // таймер для автоматического пополнения бакета
	mu       sync.Mutex   // мьютекс для защитв
	tokens   int          // текущее количество токенов в bucket
}

// NewBucket создает новый bucket с заданными настройками
func NewBucket(capacity, rate int) *Bucket {
	b := &Bucket{
		capacity: capacity,
		rate:     rate,
		tokens:   capacity,
	}
	b.ticker = time.NewTicker(time.Duration(rate) * time.Second)
	go b.refillTokens()
	return b
}

// refillTokens периодически пополняет токены в bucket
func (b *Bucket) refillTokens() {
	for range b.ticker.C {
		b.mu.Lock()
		if b.tokens < b.capacity {
			b.tokens++
		}
		b.mu.Unlock()
	}
}

// TakeToken извлекает токен из bucket, если он доступен
func (b *Bucket) TakeToken() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

// Client представляет собой клиента с bucket токенов
type Client struct {
	ip     string
	bucket *Bucket
}

// NewClient создает нового клиента с bucket токенов
func NewClient(ip string, capacity, rate int) *Client {
	return &Client{
		ip:     ip,
		bucket: NewBucket(capacity, rate),
	}
}

// TakeToken извлекает токен из bucket клиента, если он доступен
func (c *Client) TakeToken() bool {
	return c.bucket.TakeToken()
}

// RateLimiter представляет собой модуль rate-limiting
type RateLimiter struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

// NewRateLimiter создает новый модуль rate-limiting
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*Client),
	}
}

// AddClient добавляет нового клиента в модуль rate-limiting
func (r *RateLimiter) AddClient(ip string, capacity, rate int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[ip] = NewClient(ip, capacity, rate)
}

// TakeToken извлекает токен из bucket клиента, если он доступен
func (r *RateLimiter) TakeToken(ip string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	client, ok := r.clients[ip]
	if !ok {
		return false
	}
	return client.TakeToken()
}
