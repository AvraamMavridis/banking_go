package middleware

import (
	"encoding/json"
	"net/http"
)

func RequireAPIKey(expectedKey string) func(http.Handler) http.Handler {
	return func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" || key != expectedKey {
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
