package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, statusCode int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if statusCode > 499 {
		log.Printf("Internal server 5XX error: %s", msg)
	}
	type errorResp struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, statusCode, errorResp{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error marshalling JSON response", err)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(jsonResp)
}
