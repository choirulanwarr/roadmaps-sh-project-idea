package handlers

import (
	"net/http"
	"weather-api/services"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns API health status
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "Weather API is running",
	})
}

// GetWeather handles weather data requests
func GetWeather(c *gin.Context) {
	// Get query parameters
	city := c.Query("city")
	units := c.DefaultQuery("units", "metric") // metric, imperial

	if city == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "City parameter is required",
		})
		return
	}

	// Get weather data (with caching)
	weatherData, err := services.GetWeatherData(city, units)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch weather data",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, weatherData)
}
