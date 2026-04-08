package apperrors

import (
	"encoding/json"
	"net/http"
)

type AccountNotFound struct {
	Message string
}

func (e *AccountNotFound) Error() string {
	if e.Message == "" {
		return "Not Found"
	}
	return e.Message
}

func (e *AccountNotFound) StatusCode() int {
	return http.StatusNotFound
}

type BadRequest struct {
	Message string
}

func (e *BadRequest) Error() string {
	if e.Message == "" {
		return "Bad Request"
	}
	return e.Message
}

func (e *BadRequest) StatusCode() int {
	return http.StatusBadRequest
}

type InsufficientFunds struct {
	Message string
}

func (e *InsufficientFunds) Error() string {
	if e.Message == "" {
		return "Insufficient funds"
	}
	return e.Message
}

func (e *InsufficientFunds) StatusCode() int {
	return http.StatusUnprocessableEntity
}

type DuplicateRequest struct {
	StatusCode     int
	CachedResponse json.RawMessage
}

func (e *DuplicateRequest) Error() string {
	return "Duplicate request"
}

type IdempotencyKeyReused struct{}

func (e *IdempotencyKeyReused) Error() string {
	return "Idempotency key already used for a different request"
}

func (e *IdempotencyKeyReused) StatusCode() int {
	return http.StatusUnprocessableEntity
}
