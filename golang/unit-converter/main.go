package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type ConvertRequest struct {
	Amount         float64 `json:"amount"`
	FromUnit       string  `json:"fromUnit"`
	ToUnit         string  `json:"toUnit"`
	ConversionType string  `json:"conversionType"`
}

type ConvertResponse struct {
	Result float64 `json:"result"`
}

func main() {
	http.HandleFunc("/convert", handleConvert)

	fmt.Println("Server starting on http://localhost:8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := calculateConversion(req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ConvertResponse{Result: result})
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func calculateConversion(req ConvertRequest) float64 {
	if req.FromUnit == req.ToUnit {
		return req.Amount
	}

	switch req.ConversionType {
	case "length":
		return convertLength(req.Amount, req.FromUnit, req.ToUnit)
	case "weight":
		return convertWeight(req.Amount, req.FromUnit, req.ToUnit)
	case "temperature":
		return convertTemperature(req.Amount, req.FromUnit, req.ToUnit)
	default:
		return 0
	}
}

func convertLength(val float64, from, to string) float64 {
	toMeters := map[string]float64{
		"m":  1,
		"km": 1000,
		"cm": 0.01,
		"mm": 0.001,
		"ft": 0.3048,
		"in": 0.0254,
		"mi": 1609.34,
	}
	return (val * toMeters[from]) / toMeters[to]
}

func convertWeight(val float64, from, to string) float64 {
	toGrams := map[string]float64{
		"g":  1,
		"kg": 1000,
		"oz": 28.3495,
		"lb": 453.592,
	}
	return (val * toGrams[from]) / toGrams[to]
}

func convertTemperature(val float64, from, to string) float64 {
	var celsius float64

	switch from {
	case "c":
		celsius = val
	case "f":
		celsius = (val - 32) * 5 / 9
	case "k":
		celsius = val - 273.15
	}

	switch to {
	case "c":
		return celsius
	case "f":
		return (celsius * 9 / 5) + 32
	case "k":
		return celsius + 273.15
	default:
		return celsius
	}
}
