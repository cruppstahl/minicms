package core

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple rate limiting middleware
type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*Client
	limit    int
	window   time.Duration
	cleanupC chan struct{}
}

type Client struct {
	requests []time.Time
	blocked  time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		clients:  make(map[string]*Client),
		limit:    requestsPerMinute,
		window:   time.Minute,
		cleanupC: make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Middleware returns a Gin middleware function
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !rl.Allow(clientIP) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			return
		}

		c.Next()
	}
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	RecordRateLimitHit()

	now := time.Now()
	client, exists := rl.clients[clientIP]
	if !exists {
		client = &Client{
			requests: make([]time.Time, 0),
		}
		rl.clients[clientIP] = client
	}

	// Check if client is temporarily blocked
	if now.Before(client.blocked) {
		RecordRateLimitBlock()
		return false
	}

	// Remove old requests outside the window
	cutoff := now.Add(-rl.window)
	validRequests := make([]time.Time, 0)
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	client.requests = validRequests

	// Check if limit exceeded
	if len(client.requests) >= rl.limit {
		// Block for the remaining time window
		client.blocked = now.Add(rl.window)
		RecordRateLimitBlock()
		return false
	}

	// Add current request
	client.requests = append(client.requests, now)
	return true
}

// cleanup removes old client entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			cutoff := now.Add(-rl.window * 2)

			for ip, client := range rl.clients {
				// Remove clients with no recent requests
				if len(client.requests) == 0 || (len(client.requests) > 0 && client.requests[len(client.requests)-1].Before(cutoff)) {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.cleanupC:
			return
		}
	}
}

// Stop stops the rate limiter cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.cleanupC)
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent content sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent page rendering in frames (clickjacking protection)
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Enforce HTTPS (if serving over HTTPS)
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Prevent referrer leakage
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy - basic policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		c.Next()
	}
}