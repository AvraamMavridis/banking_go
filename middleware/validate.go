package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const IdempotencyKeyCtx contextKey = "idempotencyKey"

func RequireIdempotencyKey(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if _, err := uuid.Parse(key); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"statusCode": 400,
				"error":      "Bad Request",
				"message":    "Valid Idempotency-Key UUID header required",
			})
			return
		}
		ctx := context.WithValue(r.Context(), IdempotencyKeyCtx, key)
		inner.ServeHTTP(w, r.WithContext(ctx))
	})
}
