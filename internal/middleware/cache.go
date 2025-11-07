package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Over-knight/Lujay-assesment/internal/cache"
)

// CacheMiddleware creates a middleware that caches GET requests
func CacheMiddleware(redisCache *cache.RedisCache, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Generate cache key from URL and query parameters
		cacheKey := generateCacheKey(c)

		// Try to get from cache
		var cachedResponse CachedResponse
		err := redisCache.Get(c.Request.Context(), cacheKey, &cachedResponse)
		if err == nil {
			// Cache hit - return cached response
			for key, values := range cachedResponse.Headers {
				for _, value := range values {
					c.Header(key, value)
				}
			}
			c.Header("X-Cache", "HIT")
			c.Data(cachedResponse.StatusCode, cachedResponse.ContentType, cachedResponse.Body)
			c.Abort()
			return
		}

		// Cache miss - continue with request
		c.Header("X-Cache", "MISS")

		// Create a response writer wrapper to capture the response
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Only cache successful responses (2xx status codes)
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// Store response in cache
			cachedResponse = CachedResponse{
				StatusCode:  c.Writer.Status(),
				ContentType: c.Writer.Header().Get("Content-Type"),
				Headers:     c.Writer.Header(),
				Body:        responseWriter.body.Bytes(),
			}

			// Store in Redis with TTL
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = redisCache.Set(ctx, cacheKey, cachedResponse, ttl)
		}
	}
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode  int                 `json:"status_code"`
	ContentType string              `json:"content_type"`
	Headers     map[string][]string `json:"headers"`
	Body        []byte              `json:"body"`
}

// responseBodyWriter wraps gin.ResponseWriter to capture response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseBodyWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// generateCacheKey creates a unique cache key from the request
func generateCacheKey(c *gin.Context) string {
	// Include path, query params, and user ID (if authenticated)
	userID := ""
	if id, exists := c.Get("userID"); exists {
		userID = fmt.Sprintf("%v", id)
	}

	// Create a unique string from request details
	keyString := fmt.Sprintf("%s:%s:%s:%s",
		c.Request.Method,
		c.Request.URL.Path,
		c.Request.URL.RawQuery,
		userID,
	)

	// Hash the key to keep it short
	hash := sha256.Sum256([]byte(keyString))
	return "cache:" + hex.EncodeToString(hash[:])
}

// InvalidateCache invalidates cache for specific patterns
func InvalidateCache(redisCache *cache.RedisCache, pattern string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// If request was successful, invalidate related cache
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			keys, err := redisCache.Keys(ctx, pattern)
			if err == nil && len(keys) > 0 {
				_ = redisCache.Delete(ctx, keys...)
			}
		}
	}
}

// CacheBuster middleware that clears cache on POST, PUT, PATCH, DELETE requests
func CacheBuster(redisCache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store original method
		method := c.Request.Method

		c.Next()

		// If it was a modifying request and successful, clear relevant cache
		if (method == http.MethodPost || method == http.MethodPut ||
			method == http.MethodPatch || method == http.MethodDelete) &&
			c.Writer.Status() >= 200 && c.Writer.Status() < 300 {

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Clear cache for this resource path
			basePath := c.Request.URL.Path
			pattern := fmt.Sprintf("cache:*%s*", basePath)

			keys, err := redisCache.Keys(ctx, pattern)
			if err == nil && len(keys) > 0 {
				_ = redisCache.Delete(ctx, keys...)
			}
		}
	}
}

// RateLimitMiddleware implements Redis-based rate limiting
func RateLimitMiddleware(redisCache *cache.RedisCache, maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP or user ID)
		identifier := c.ClientIP()
		if userID, exists := c.Get("userID"); exists {
			identifier = fmt.Sprintf("user:%v", userID)
		}

		// Create rate limit key
		key := fmt.Sprintf("ratelimit:%s", identifier)

		ctx := c.Request.Context()

		// Check if key exists
		exists, err := redisCache.Exists(ctx, key)
		if err != nil {
			c.Next()
			return
		}

		if !exists {
			// First request in this window
			if err := redisCache.Set(ctx, key, 1, window); err != nil {
				c.Next()
				return
			}
			c.Next()
			return
		}

		// Increment counter
		count, err := redisCache.Increment(ctx, key)
		if err != nil {
			c.Next()
			return
		}

		// Set expiration if this is the first increment
		if count == 2 {
			_ = redisCache.Expire(ctx, key, window)
		}

		// Check if limit exceeded
		if count > int64(maxRequests) {
			// Get TTL for Retry-After header
			ttl, _ := redisCache.TTL(ctx, key)
			c.Header("Retry-After", fmt.Sprintf("%.0f", ttl.Seconds()))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":               "Rate limit exceeded",
				"retry_after_seconds": int(ttl.Seconds()),
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", maxRequests-int(count)))

		c.Next()
	}
}

// SessionCache middleware for caching user sessions
func SessionCacheMiddleware(redisCache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		sessionKey := fmt.Sprintf("session:%v", userID)

		// Try to get user data from cache
		var userData map[string]interface{}
		err := redisCache.Get(c.Request.Context(), sessionKey, &userData)
		if err == nil {
			// Session found in cache
			c.Set("cachedUserData", userData)
		}

		c.Next()
	}
}

// GetFromCache is a helper to get data from cache in handlers
func GetFromCache(c *gin.Context, redisCache *cache.RedisCache, key string, dest interface{}) error {
	return redisCache.Get(c.Request.Context(), key, dest)
}

// SetInCache is a helper to set data in cache from handlers
func SetInCache(c *gin.Context, redisCache *cache.RedisCache, key string, value interface{}, ttl time.Duration) error {
	return redisCache.Set(c.Request.Context(), key, value, ttl)
}

// ClearCache is a helper to clear cache from handlers
func ClearCache(c *gin.Context, redisCache *cache.RedisCache, pattern string) error {
	ctx := c.Request.Context()
	keys, err := redisCache.Keys(ctx, pattern)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return redisCache.Delete(ctx, keys...)
	}
	return nil
}
