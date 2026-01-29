package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/mr_rambling/chirpy/internal/auth"
	"net/http"
)

func (cfg *apiConfig) handlerWebhooks(w http.ResponseWriter, r *http.Request) {
	type webhookPayload struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	var payload webhookPayload
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong decoding the JSON body", err)
		return
	}

	if payload.Event != "user.upgraded" {
		respondWithError(w, http.StatusNoContent, "Unsupported event type", nil)
		return
	}

	err = cfg.db.UpgradeChirpyRed(r.Context(), payload.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Something went wrong upgrading the user to premium", err)
		return
	}

	keyStr, err := auth.GetAPIKey(r.Header)
	if err != nil || keyStr != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized user", err)
	}

	respondWithJSON(w, http.StatusNoContent, "")
}
