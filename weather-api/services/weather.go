package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// WeatherAPIResponse represents the response from Visual Crossing API
type WeatherAPIResponse struct {
	QueryCost         int               `json:"queryCost"`
	Longitude         float64           `json:"longitude"`
	Latitude          float64           `json:"latitude"`
	ResolvedAddress   string            `json:"resolvedAddress"`
	Address           string            `json:"address"`
	Timezone          string            `json:"timezone"`
	Tzoffset          float64           `json:"tzoffset"`
	Days              []Day             `json:"days"`
	CurrentConditions CurrentConditions `json:"currentConditions"`
}

type Day struct {
	Datetime    string  `json:"datetime"`
	Temp        float64 `json:"temp"`
	Feelslike   float64 `json:"feelslike"`
	Humidity    float64 `json:"humidity"`
	Windspeed   float64 `json:"windspeed"`
	Winddir     float64 `json:"winddir"`
	Pressure    float64 `json:"pressure"`
	Visibility  float64 `json:"visibility"`
	Cloudcover  float64 `json:"cloudcover"`
	Precip      float64 `json:"precip"`
	Precipprob  float64 `json:"precipprob"`
	Snow        float64 `json:"snow"`
	Snowdepth   float64 `json:"snowdepth"`
	Conditions  string  `json:"conditions"`
	Description string  `json:"description"`
}

type CurrentConditions struct {
	Datetime    string  `json:"datetime"`
	Temp        float64 `json:"temp"`
	Feelslike   float64 `json:"feelslike"`
	Humidity    float64 `json:"humidity"`
	Windspeed   float64 `json:"windspeed"`
	Winddir     float64 `json:"winddir"`
	Pressure    float64 `json:"pressure"`
	Visibility  float64 `json:"visibility"`
	Cloudcover  float64 `json:"cloudcover"`
	Precip      float64 `json:"precip"`
	Precipprob  float64 `json:"precipprob"`
	Snow        float64 `json:"snow"`
	Snowdepth   float64 `json:"snowdepth"`
	Conditions  string  `json:"conditions"`
	Description string  `json:"description"`
}

// GetWeatherData fetches weather data with caching
func GetWeatherData(city, units string) (*WeatherAPIResponse, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("weather:%s:%s", strings.ToLower(city), units)

	// Try to get from cache first
	var cachedData WeatherAPIResponse
	if GetFromCache(cacheKey, &cachedData) {
		return &cachedData, nil
	}

	// Fetch from 3rd party API
	apiData, err := fetchFromAPI(city, units)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := SetCache(cacheKey, apiData); err != nil {
		log.Printf("Failed to cache data: %v", err)
	}

	return apiData, nil
}

// fetchFromAPI calls the Visual Crossing Weather API
func fetchFromAPI(city, units string) (*WeatherAPIResponse, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("WEATHER_API_KEY not set in environment variables")
	}

	baseURL := "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/"

	// Build URL
	encodedCity := url.QueryEscape(city)
	unitSystem := "metric"
	if units == "imperial" {
		unitSystem = "us"
	}

	apiURL := fmt.Sprintf("%s%s?unitGroup=%s&key=%s&contentType=json",
		baseURL, encodedCity, unitSystem, apiKey)

	// Make HTTP request
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var weatherData WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	log.Printf("🌍 Fetched weather data for %s from API", city)
	return &weatherData, nil
}
