package ratelimit

import (
	"sync"
	"time"
)

// Bucket представляет собой bucket токенов для клиента
type Bucket struct {
	capacity int // емкость bucket
	rate     int // скорость пополнения bucket
	tokens   int // текущее количество токенов в bucket
	mu       sync.Mutex
	ticker   *time.Ticker
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

// ReturnToken возвращает токен обратно в bucket
func (b *Bucket) ReturnToken() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.tokens < b.capacity {
		b.tokens++
	}
	// Не превышаем максимальную емкость
}
