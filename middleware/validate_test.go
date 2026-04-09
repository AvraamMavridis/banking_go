package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireIdempotencyKey_ValidUUID(t *testing.T) {
	var capturedKey string
	handler := RequireIdempotencyKey(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedKey, _ = r.Context().Value(IdempotencyKeyCtx).(string)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Idempotency-Key", "550e8400-e29b-41d4-a716-446655440000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if capturedKey != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("expected key in context, got %q", capturedKey)
	}
}

func TestRequireIdempotencyKey_MissingHeader(t *testing.T) {
	handler := RequireIdempotencyKey(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["message"] != "Valid Idempotency-Key UUID header required" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}

func TestRequireIdempotencyKey_InvalidUUID(t *testing.T) {
	handler := RequireIdempotencyKey(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Idempotency-Key", "not-a-uuid")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
