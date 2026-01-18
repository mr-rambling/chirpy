package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/mr_rambling/chirpy/internal/auth"
	"github.com/mr_rambling/chirpy/internal/database"
	"net/http"
	"time"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	AccToken  string    `json:"token,omitempty"`
	RefToken  string    `json:"refresh_token"`
}

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
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var params parameters
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong decoding the JSON body", err)
		return
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

	token, err := auth.MakeJWT(dbUser.ID, cfg.secretKey, time.Duration(3600)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong creating the JWT token", err)
		return
	}

	refToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong creating the refresh token", err)
		return
	}

	expiry := time.Now().Add(60 * 24 * time.Hour)
	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refToken,
		UserID:    dbUser.ID,
		ExpiresAt: expiry,
	})

	u := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		AccToken:  token,
		RefToken:  refToken,
	}
	respondWithJSON(w, http.StatusOK, u)
}

func (cfg *apiConfig) handlerUserUpdate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
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

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong decoding the JSON body", err)
		return
	}

	hashedPw, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong hashing the password", err)
	}
	dbUser, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:           id,
		Email:        params.Email,
		PasswordHash: hashedPw,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong updating the user", err)
		return
	}

	u := User{
		ID:        id,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	respondWithJSON(w, http.StatusOK, u)
}
