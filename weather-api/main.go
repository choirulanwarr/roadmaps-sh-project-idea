package main

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"weather-api/handlers"
	"weather-api/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// RateLimiter struct untuk menyimpan state per IP
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Cleanup lama secara berkala
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, timestamps := range rl.visitors {
		// Hapus timestamp yang sudah melewati window
		newTimestamps := []time.Time{}
		for _, t := range timestamps {
			if now.Sub(t) <= rl.window {
				newTimestamps = append(newTimestamps, t)
			}
		}
		if len(newTimestamps) == 0 {
			delete(rl.visitors, ip)
		} else {
			rl.visitors[ip] = newTimestamps
		}
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	timestamps := rl.visitors[ip]

	// Filter timestamp yang masih dalam window
	validTimestamps := []time.Time{}
	for _, t := range timestamps {
		if now.Sub(t) <= rl.window {
			validTimestamps = append(validTimestamps, t)
		}
	}

	// Jika sudah melebihi limit → tolak
	if len(validTimestamps) >= rl.limit {
		return false
	}

	// Catat request baru
	rl.visitors[ip] = append(validTimestamps, now)
	return true
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables")
	}

	// Initialize Redis connection
	if err := services.InitRedis(); err != nil {
		log.Printf("⚠️  Redis initialization failed (running without cache): %v", err)
		// Tetap lanjut tanpa cache (graceful degradation)
	}

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Serve static files dari folder frontend
	r.Static("/css", "./frontend/css")
	r.Static("/js", "./frontend/js")

	// Rate limiting: 100 requests per hour per IP
	rateLimiter := NewRateLimiter(100, 1*time.Hour)
	r.Use(func(c *gin.Context) {
		ip := c.ClientIP()
		if !rateLimiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Rate limit exceeded. Try again later.",
				"limit":   "100 requests per hour",
			})
			c.Abort()
			return
		}
		c.Next()
	})

	// API Routes
	api := r.Group("/api")
	{
		api.GET("/weather", handlers.GetWeather)
		api.GET("/health", handlers.HealthCheck)
	}

	// Serve index.html untuk root path
	r.GET("/", func(c *gin.Context) {
		c.File("./frontend/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
