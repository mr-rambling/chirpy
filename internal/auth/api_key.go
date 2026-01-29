package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	keyStr := headers.Get("Authorization")

	if keyStr == "" {
		return keyStr, fmt.Errorf("Authorization header not found")
	}
	keyStr = strings.ReplaceAll(keyStr, "ApiKey ", "")
	return keyStr, nil
}
