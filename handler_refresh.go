package main

import (
	"net/http"
	"time"

	"github.com/mr_rambling/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerTokenRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization token missing", err)
		return
	}

	refToken, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil || refToken.ExpiresAt.Before(time.Now()) || refToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	newToken, err := auth.MakeJWT(refToken.UserID, cfg.secretKey, time.Duration(3600)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong creating the JWT token", err)
		return
	}

	type tokenResponse struct {
		AccessToken string `json:"token"`
	}

	respondWithJSON(w, http.StatusOK, tokenResponse{
		AccessToken: newToken,
	})
}
