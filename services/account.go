package services

import (
	"bank_api_go/apperrors"
	"bank_api_go/entities"

	"gorm.io/gorm"
)

type AccountService struct {
	db          *gorm.DB
	idempotency *IdempotencyService
}

func NewAccountService(db *gorm.DB, idempotency *IdempotencyService) *AccountService {
	return &AccountService{db: db, idempotency: idempotency}
}

func (s *AccountService) FindByID(id uint) (*entities.Account, error) {
	var account entities.Account
	if err := s.db.First(&account, id).Error; err != nil {
		return nil, &apperrors.AccountNotFound{Message: "Account not found"}
	}
	return &account, nil
}

func (s *AccountService) Create(idempotencyKey string, fingerprint string, account *entities.Account) (*entities.Account, error) {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.idempotency.EnsureUnique(tx, idempotencyKey, fingerprint); err != nil {
			return err
		}

		if err := tx.Create(account).Error; err != nil {
			return err
		}

		return s.idempotency.Save(tx, idempotencyKey, fingerprint, 201, account)
	})
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (s *AccountService) Deposit(idempotencyKey string, fingerprint string, id uint, amount int64) (*entities.Account, error) {
	var account entities.Account
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.idempotency.EnsureUnique(tx, idempotencyKey, fingerprint); err != nil {
			return err
		}

		if err := tx.First(&account, id).Error; err != nil {
			return &apperrors.AccountNotFound{Message: "Account not found"}
		}

		if err := tx.Model(&account).Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}
		account.Balance += amount

		return s.idempotency.Save(tx, idempotencyKey, fingerprint, 200, &account)
	})
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (s *AccountService) Transfer(idempotencyKey string, fingerprint string, fromID, toID uint, amount int64) (*entities.TransferResult, error) {
	if fromID == toID {
		return nil, &apperrors.BadRequest{Message: "Cannot transfer to the same account"}
	}

	var result entities.TransferResult
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.idempotency.EnsureUnique(tx, idempotencyKey, fingerprint); err != nil {
			return err
		}

		if err := tx.First(&result.From, fromID).Error; err != nil {
			return &apperrors.AccountNotFound{Message: "Source account not found"}
		}
		if err := tx.First(&result.To, toID).Error; err != nil {
			return &apperrors.AccountNotFound{Message: "Destination account not found"}
		}

		if result.From.Currency != result.To.Currency {
			return &apperrors.BadRequest{Message: "Cannot transfer between accounts with different currencies"}
		}

		if result.From.Balance < amount {
			return &apperrors.InsufficientFunds{}
		}

		if err := tx.Model(&result.From).Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
			return err
		}
		result.From.Balance -= amount

		if err := tx.Model(&result.To).Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}
		result.To.Balance += amount

		return s.idempotency.Save(tx, idempotencyKey, fingerprint, 200, &result)
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}
