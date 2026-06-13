package main

import (
	"net/http"
	"encoding/json"
	"log"
	"time"
)

type Response struct {
	Nome	  string `json:"nome"`
	Horario   string `json:"horario"`
}

func projetoHandler(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("GET /projeto-korp", projetoHandler)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}