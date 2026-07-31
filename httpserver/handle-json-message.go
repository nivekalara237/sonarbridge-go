package main

import (
	"encoding/json"
	"net/http"
)

type Message struct {
	Name string `json:"name"`
	Body string `json:"body"`
}

func handleJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg Message

	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	response := map[string]string {
		"status": "success",
		"message": "Received message from " + msg.Name,
		"body": msg.Body,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}
