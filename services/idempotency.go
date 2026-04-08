package services

import (
	"encoding/json"

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
	if err := tx.Where("key = ?", key).First(&record).Error; err == nil {
		if record.Fingerprint != fingerprint {
			return &apperrors.IdempotencyKeyReused{}
		}
		return &apperrors.DuplicateRequest{
			StatusCode:     record.StatusCode,
			CachedResponse: json.RawMessage(record.Response),
		}
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
