package main

import (
	"github.com/mr_rambling/chirpy/internal/auth"
	"net/http"
)

func (cfg *apiConfig) handlerTokenRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization token missing", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong revoking the refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
