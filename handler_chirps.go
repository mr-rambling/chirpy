package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/mr_rambling/chirpy/internal/auth"
	"github.com/mr_rambling/chirpy/internal/database"
)

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong decoding the JSON body", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization token missing", err)
		return
	}

	id, err := auth.ValidateJWT(token, cfg.secretKey)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization token", err)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	type returnVal struct {
		Censored string `json:"cleaned_body"`
	}

	cleaned := returnVal{Censored: censorChirp(params.Body)}

	dbChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned.Censored,
		UserID: id,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong creating the chirp", err)
		return
	}

	c := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, c)
}

func censorChirp(chirp string) string {
	badwords := []string{"kerfuffle", "sharbert", "fornax"}
	msg := strings.Split(chirp, " ")
	for i, word := range msg {
		if slices.Contains(badwords, strings.ToLower(word)) {
			msg[i] = "****"
		}
	}
	return strings.Join(msg, " ")
}

func (cfg *apiConfig) handlerRetrieveChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong retrieving the chirps", err)
		return
	}

	var chirps []Chirp
	for _, chirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerRetrieveChirp(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("chirpID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	dbChirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Something went wrong retreiving the chirp", err)
		return
	}

	if dbChirp.ID == uuid.Nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", nil)
		return
	}

	c := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, c)
}
