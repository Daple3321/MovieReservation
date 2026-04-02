package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

func WriteJSONResponse(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing body request")
	}

	return json.NewDecoder(r.Body).Decode(payload)
}

func CheckIfFileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {

		return true
	}

	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return false
}
