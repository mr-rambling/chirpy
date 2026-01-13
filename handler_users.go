package main

import (
	"encoding/json"
	"github.com/mr_rambling/chirpy/internal/auth"
	"github.com/mr_rambling/chirpy/internal/database"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var params parameters
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong decoding the JSON body", err)
	}

	hashedPw, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong hashing the password", err)
	}
	dbUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, PasswordHash: hashedPw})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong creating the user", err)
	}

	u := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	respondWithJSON(w, http.StatusCreated, u)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds *int   `json:"expires_in_seconds"`
	}

	var params parameters
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong decoding the JSON body", err)
		return
	}

	if params.ExpiresInSeconds == nil || *params.ExpiresInSeconds >= 3600 {
		defaultExpiry := 3600 // 1 hour
		params.ExpiresInSeconds = &defaultExpiry
	}

	dbUser, err := cfg.db.GetUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, dbUser.PasswordHash)
	if err != nil || match != true {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	token, err := auth.MakeJWT(dbUser.ID, cfg.secretKey, time.Duration(*params.ExpiresInSeconds)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong creating the JWT token", err)
		return
	}

	u := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Token:     token,
	}
	respondWithJSON(w, http.StatusOK, u)
}
