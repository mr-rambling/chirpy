package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	tokenStr := headers.Get("Authorization")
	if tokenStr == "" {
		return tokenStr, fmt.Errorf("Authorization header does not exist")
	}
	tokenStr = strings.TrimSpace(strings.TrimPrefix(tokenStr, "Bearer"))
	return tokenStr, nil
}
