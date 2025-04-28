package ratelimit

import (
	"sync"
	"time"
)

// Client представляет собой клиента с bucket токенов
type Client struct {
	ip       string       // адрес клиента
	bucket   *Bucket      // бакет клиента
	mu       sync.RWMutex // мьютекс защиты данных (lastSeen)
	lastSeen time.Time    // время последней активности клиента
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

func (c *Client) ReturnToken() {
	c.bucket.ReturnToken()
}

// Добавляем метод обновления времени последней активности
func (c *Client) UpdateLastSeen() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastSeen = time.Now()
}

// Добавляем метод проверки активности
func (c *Client) IsActive(timeout time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.lastSeen) < timeout
}
