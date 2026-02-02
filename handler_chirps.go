package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/mr_rambling/chirpy/internal/auth"
	"github.com/mr_rambling/chirpy/internal/database"
	"net/http"
	"slices"
	"sort"
	"strings"
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
	authorID := uuid.Nil
	authStr := r.URL.Query().Get("author_id")
	authorID, err := uuid.Parse(authStr)
	if err != nil && authStr != "" {
		respondWithError(w, http.StatusBadRequest, "Invalid author ID", err)
		return
	}

	sortQ := r.URL.Query().Get("sort")
	if sortQ == "" {
		sortQ = "asc"
	} else if sortQ != "asc" && sortQ != "desc" {
		respondWithError(w, http.StatusBadRequest, "Invalid sort query", nil)
		return
	}

	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong retrieving the chirps", err)
		return
	}

	var chirps []Chirp
	for _, chirp := range dbChirps {
		if authorID != uuid.Nil && chirp.UserID != authorID {
			continue
		}
		chirps = append(chirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	if sortQ == "asc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
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

func (cfg *apiConfig) handlerChirpDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("chirpID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization token missing", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secretKey)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization token", err)
		return
	}

	dbChirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong retrieving the chirp", err)
		return
	}

	if dbChirp.ID == uuid.Nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", nil)
		return
	}

	if dbChirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You do not have permission to delete this chirp", nil)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong deleting the chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
