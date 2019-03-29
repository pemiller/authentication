package helpers

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

// Authorization Types
const (
	AuthTypeToken = "Token"
	AuthTypeCode  = "Code"
	AuthTypeBasic = "Basic"
)

// ParseAuthorizationHeader returns the value of the authorization header if it matches the type
func ParseAuthorizationHeader(r *http.Request, authType string) (string, error) {
	headerValue := r.Header.Get("Authorization")
	l := len(authType)
	if len(headerValue) > l+1 && headerValue[:l] == authType {
		return headerValue[l+1:], nil
	}

	return "", errors.New("Unable to find authorization header with type " + authType)
}

// DecodeBasicCredentials decodes username and password values from Basic Authorization
func DecodeBasicCredentials(headerValue string) (string, string, error) {
	bytes, err := base64.StdEncoding.DecodeString(headerValue)
	if err != nil {
		return "", "", err
	}

	credentials := string(bytes)
	parts := strings.Split(credentials, ":")
	if len(parts) < 2 {
		return parts[0], "", nil
	}

	return parts[0], strings.Join(parts[1:], ":"), nil
}
