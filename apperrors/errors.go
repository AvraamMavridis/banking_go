package apperrors

import (
	"encoding/json"
	"net/http"
)

type AccountNotFound struct {
	Message string
}

func (err *AccountNotFound) Error() string {
	if err.Message == "" {
		return "Not Found"
	}
	return err.Message
}

func (err *AccountNotFound) StatusCode() int {
	return http.StatusNotFound
}

type BadRequest struct {
	Message string
}

func (err *BadRequest) Error() string {
	if err.Message == "" {
		return "Bad Request"
	}
	return err.Message
}

func (err *BadRequest) StatusCode() int {
	return http.StatusBadRequest
}

type InsufficientFunds struct {
	Message string
}

func (err *InsufficientFunds) Error() string {
	if err.Message == "" {
		return "Insufficient funds"
	}
	return err.Message
}

func (err *InsufficientFunds) StatusCode() int {
	return http.StatusUnprocessableEntity
}

type DuplicateRequest struct {
	CachedStatusCode int
	CachedResponse   json.RawMessage
}

func (err *DuplicateRequest) Error() string {
	return "Duplicate request"
}

func (err *DuplicateRequest) StatusCode() int {
	return err.CachedStatusCode
}

func (err *DuplicateRequest) Response() map[string]any {
	var data any
	_ = json.Unmarshal(err.CachedResponse, &data)
	return map[string]any{"cachedResponse": data}
}

type IdempotencyKeyReused struct{}

func (err *IdempotencyKeyReused) Error() string {
	return "Idempotency key already used for a different request"
}

func (err *IdempotencyKeyReused) StatusCode() int {
	return http.StatusUnprocessableEntity
}
