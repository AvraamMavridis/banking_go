package services

import (
	"encoding/json"
	"errors"

	"bank_api_go/apperrors"
	"bank_api_go/entities"

	"gorm.io/gorm"
)

type IdempotencyService struct{}

func NewIdempotencyService() *IdempotencyService {
	return &IdempotencyService{}
}

func (s *IdempotencyService) EnsureUnique(tx *gorm.DB, key string, fingerprint string) error {
	var record entities.IdempotencyRecord
	err := tx.Where("key = ?", key).First(&record).Error
	if err == nil {
		if record.Fingerprint != fingerprint {
			return &apperrors.IdempotencyKeyReused{}
		}
		return &apperrors.DuplicateRequest{
			CachedStatusCode: record.StatusCode,
			CachedResponse:   json.RawMessage(record.Response),
		}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

func (s *IdempotencyService) Save(tx *gorm.DB, key string, fingerprint string, statusCode int, response any) error {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	record := entities.IdempotencyRecord{
		Key:         key,
		Fingerprint: fingerprint,
		Response:    string(data),
		StatusCode:  statusCode,
	}
	return tx.Create(&record).Error
}
