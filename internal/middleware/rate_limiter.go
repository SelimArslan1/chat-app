package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int
	window   time.Duration
}

type visitor struct {
	count    int
	lastSeen time.Time
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) isAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	if !exists {
		rl.visitors[key] = &visitor{count: 1, lastSeen: time.Now()}
		return true
	}

	if time.Since(v.lastSeen) > rl.window {
		v.count = 1
		v.lastSeen = time.Now()
		return true
	}

	if v.count >= rl.rate {
		return false
	}

	v.count++
	v.lastSeen = time.Now()
	return true
}

func RateLimitByIP(rate int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, window)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.isAllowed(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			return
		}

		c.Next()
	}
}

func RateLimitByUser(rate int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, window)

	return func(c *gin.Context) {
		// Try to get user_id, fall back to IP
		key := c.GetString("user_id")
		if key == "" {
			key = c.ClientIP()
		}

		if !limiter.isAllowed(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			return
		}

		c.Next()
	}
}

// StrictRateLimit creates a very strict limiter for sensitive endpoints (auth)
func StrictRateLimit() gin.HandlerFunc {
	return RateLimitByIP(5, time.Minute) // 5 requests per minute
}

// StandardRateLimit creates a standard limiter for API endpoints
func StandardRateLimit() gin.HandlerFunc {
	return RateLimitByIP(60, time.Minute) // 60 requests per minute
}

// MessageRateLimit creates a limiter for message sending
func MessageRateLimit() gin.HandlerFunc {
	return RateLimitByUser(30, time.Minute) // 30 messages per minute
}
