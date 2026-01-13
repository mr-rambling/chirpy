package auth

import (
	"github.com/google/uuid"
	"net/http"
	"testing"
	"time"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "my_secure_password_123!"

	hashedPw, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}

	match, err := CheckPasswordHash(password, hashedPw)
	if err != nil {
		t.Fatalf("Error checking password hash: %v", err)
	}
	if !match {
		t.Fatalf("Password did not match the hash")
	}

	wrongPassword := "wrong_password"
	match, err = CheckPasswordHash(wrongPassword, hashedPw)
	if err != nil {
		t.Fatalf("Error checking password hash with wrong password: %v", err)
	}
	if match {
		t.Fatalf("Wrong password matched the hash")
	}
}

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "my_secret_key"
	expiresIn := time.Minute * 15

	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("Error making JWT: %v", err)
	}

	returnedUserID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("Error validating JWT: %v", err)
	}
	if returnedUserID != userID {
		t.Fatalf("Returned user ID does not match original. Got %v, want %v", returnedUserID, userID)
	}

	// Test with an invalid token
	_, err = ValidateJWT(token+"invalid", tokenSecret)
	if err == nil {
		t.Fatalf("Expected error when validating invalid token, got none")
	}
}

func TestValidateJWTExpiredToken(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "my_secret_key"
	expiresIn := -time.Minute * 1 // Token already expired

	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("Error making JWT: %v", err)
	}

	_, err = ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Fatalf("Expected error when validating expired token, got none")
	}
}

func TestBearerTokenExtraction(t *testing.T) {
	validHeader := http.Header{}
	validHeader.Add("Authorization", "Bearer valid_token_string")

	token, err := GetBearerToken(validHeader)
	if err != nil {
		t.Fatalf("Error extracting bearer token: %v", err)
	}
	if token != "valid_token_string" {
		t.Fatalf("Extracted token does not match expected. Got %v, want %v", token, "valid_token_string")
	}
}
