package models

// WeatherResponse is the simplified response sent to frontend
type WeatherResponse struct {
	City        string  `json:"city"`
	Country     string  `json:"country"`
	Temperature float64 `json:"temperature"`
	FeelsLike   float64 `json:"feelsLike"`
	Humidity    float64 `json:"humidity"`
	WindSpeed   float64 `json:"windSpeed"`
	Conditions  string  `json:"conditions"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
	LastUpdated string  `json:"lastUpdated"`
	Cached      bool    `json:"cached"`
}
