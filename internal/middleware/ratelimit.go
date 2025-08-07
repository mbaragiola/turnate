package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}

	// Clean up old limiters every 5 minutes
	go i.cleanupLimiters()

	return i
}

func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = limiter

	return limiter
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		return i.AddIP(ip)
	}

	return limiter
}

func (i *IPRateLimiter) cleanupLimiters() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		// Remove limiters that haven't been used recently
		for ip, limiter := range i.ips {
			if limiter.TokensAt(time.Now()) == float64(i.b) {
				delete(i.ips, ip)
			}
		}
		i.mu.Unlock()
	}
}

var (
	// Global rate limiters
	globalLimiter = NewIPRateLimiter(rate.Every(time.Second/10), 20)  // 10 requests per second, burst of 20
	authLimiter   = NewIPRateLimiter(rate.Every(time.Minute/5), 5)    // 5 login attempts per minute
	apiLimiter    = NewIPRateLimiter(rate.Every(time.Second/5), 10)   // 5 API calls per second, burst of 10
)

// RateLimitMiddleware provides general rate limiting
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := globalLimiter.GetLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please slow down.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthRateLimitMiddleware provides strict rate limiting for auth endpoints
func AuthRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := authLimiter.GetLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Authentication rate limit exceeded",
				"message": "Too many authentication attempts. Please wait before trying again.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// APIRateLimitMiddleware provides rate limiting for API endpoints
func APIRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := apiLimiter.GetLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "API rate limit exceeded",
				"message": "Too many API requests. Please slow down.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}