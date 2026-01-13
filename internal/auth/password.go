package auth

import (
	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	var params *argon2id.Params
	params = argon2id.DefaultParams

	hashedPw, err := argon2id.CreateHash(password, params)
	if err != nil {
		return "", err
	}
	return hashedPw, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	check, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return check, nil
}
