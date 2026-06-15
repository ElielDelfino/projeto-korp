package main

import (
	"net/http"
	"encoding/json"
	"log"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var httpRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "code"},
)

type Response struct {
	Nome	  string `json:"nome"`
	Horario   string `json:"horario"`
}

func projectHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Nome:    "Projeto Korp",
		Horario: time.Now().UTC().Format("15:04:05 UTC"),
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error generating response", http.StatusInternalServerError)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("GET /projeto-korp",
		promhttp.InstrumentHandlerCounter(httpRequestsTotal, http.HandlerFunc(projectHandler),
		),
	)

	mux.Handle("GET /metrics", promhttp.Handler())

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}