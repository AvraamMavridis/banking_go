package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
)

func RequireAPIKey(expectedKey string) func(http.Handler) http.Handler {
	return func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" || subtle.ConstantTimeCompare([]byte(key), []byte(expectedKey)) != 1 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]any{
					"statusCode": 401,
					"error":      "Unauthorized",
					"message":    "Invalid or missing API key",
				})
				return
			}
			inner.ServeHTTP(w, r)
		})
	}
}
