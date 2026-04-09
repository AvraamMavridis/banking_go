package services

import (
	"fmt"
	"sync"
	"testing"

	"bank_api_go/apperrors"
	"bank_api_go/entities"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&entities.Account{}, &entities.IdempotencyRecord{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestFindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAccountService(db, NewIdempotencyService())

	_, err := svc.FindByID(999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.AccountNotFound); !ok {
		t.Fatalf("expected *apperrors.AccountNotFound, got %T", err)
	}
}

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAccountService(db, NewIdempotencyService())

	account := &entities.Account{
		Name:         "John",
		Surname:      "Doe",
		Email:        "john@example.com",
		AddressLine1: "123 Main St",
		City:         "London",
		Postcode:     "SW1A 1AA",
		Country:      "UK",
		Currency:     "GBP",
	}

	created, err := svc.Create("key-1", "POST /accounts", account)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected account to have an ID")
	}
	if created.Name != "John" {
		t.Fatalf("expected name John, got %s", created.Name)
	}

	found, err := svc.FindByID(created.ID)
	if err != nil {
		t.Fatalf("unexpected error finding created account: %v", err)
	}
	if found.Email != "john@example.com" {
		t.Fatalf("expected email john@example.com, got %s", found.Email)
	}
}

func TestCreate_DuplicateIdempotencyKey(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAccountService(db, NewIdempotencyService())

	account := &entities.Account{
		Name:         "Jane",
		Surname:      "Doe",
		Email:        "jane@example.com",
		AddressLine1: "456 High St",
		City:         "London",
		Postcode:     "SW1A 2AA",
		Country:      "UK",
		Currency:     "GBP",
	}

	_, err := svc.Create("dup-key", "POST /accounts", account)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	account2 := &entities.Account{
		Name:         "Bob",
		Surname:      "Smith",
		Email:        "bob@example.com",
		AddressLine1: "789 Low St",
		City:         "London",
		Postcode:     "SW1A 3AA",
		Country:      "UK",
		Currency:     "GBP",
	}

	_, err = svc.Create("dup-key", "POST /accounts", account2)
	if err == nil {
		t.Fatal("expected error for duplicate key, got nil")
	}
	if _, ok := err.(*apperrors.DuplicateRequest); !ok {
		t.Fatalf("expected *apperrors.DuplicateRequest, got %T", err)
	}
}

func TestIdempotencyKey_ReusedAcrossEndpoints(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAccountService(db, NewIdempotencyService())

	account := &entities.Account{
		Name: "Test", Surname: "User", Email: "test@example.com",
		AddressLine1: "1 St", City: "London", Postcode: "E1", Country: "UK",
		Balance: 1000, Currency: "GBP",
	}
	_, err := svc.Create("shared-key", "POST /accounts", account)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.Deposit("shared-key", "POST /accounts/{id}/deposit", account.ID, 500)
	if err == nil {
		t.Fatal("expected error when reusing idempotency key across endpoints, got nil")
	}
	if _, ok := err.(*apperrors.IdempotencyKeyReused); !ok {
		t.Fatalf("expected *apperrors.IdempotencyKeyReused, got %T", err)
	}
}

func TestDeposit(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAccountService(db, NewIdempotencyService())

	account := &entities.Account{
		Name:         "Alice",
		Surname:      "Smith",
		Email:        "alice@example.com",
		AddressLine1: "10 Park Ave",
		City:         "London",
		Postcode:     "E1 6AN",
		Country:      "UK",
		Balance:      1000,
		Currency:     "GBP",
	}
	created, _ := svc.Create("create-key", "POST /accounts", account)

	updated, err := svc.Deposit("deposit-key", "POST /accounts/1/deposit", created.ID, 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Balance != 1500 {
		t.Fatalf("expected balance 1500, got %d", updated.Balance)
	}
}

func TestDeposit_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAccountService(db, NewIdempotencyService())

	_, err := svc.Deposit("dep-key", "POST /accounts/999/deposit", 999, 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.AccountNotFound); !ok {
		t.Fatalf("expected *apperrors.AccountNotFound, got %T", err)
	}
}

func TestTransfer(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAccountService(db, NewIdempotencyService())

	from := &entities.Account{
		Name: "From", Surname: "User", Email: "from@example.com",
		AddressLine1: "1 St", City: "London", Postcode: "E1", Country: "UK",
		Balance: 1000, Currency: "GBP",
	}
	to := &entities.Account{
		Name: "To", Surname: "User", Email: "to@example.com",
		AddressLine1: "2 St", City: "London", Postcode: "E2", Country: "UK",
		Balance: 500, Currency: "GBP",
	}

	from, _ = svc.Create("key-from", "POST /accounts", from)
	to, _ = svc.Create("key-to", "POST /accounts", to)

	result, err := svc.Transfer("transfer-key", "POST /accounts/1/transfer", from.ID, to.ID, 300)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.From.Balance != 700 {
		t.Fatalf("expected from balance 700, got %d", result.From.Balance)
	}
	if result.To.Balance != 800 {
		t.Fatalf("expected to balance 800, got %d", result.To.Balance)
	}
}

func TestTransfer_InsufficientFunds(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAccountService(db, NewIdempotencyService())

	from := &entities.Account{
		Name: "Poor", Surname: "User", Email: "poor@example.com",
		AddressLine1: "1 St", City: "London", Postcode: "E1", Country: "UK",
		Balance: 100, Currency: "GBP",
	}
	to := &entities.Account{
		Name: "Rich", Surname: "User", Email: "rich@example.com",
		AddressLine1: "2 St", City: "London", Postcode: "E2", Country: "UK",
		Balance: 500, Currency: "GBP",
	}

	from, _ = svc.Create("key-f", "POST /accounts", from)
	to, _ = svc.Create("key-t", "POST /accounts", to)

	_, err := svc.Transfer("transfer-key", "POST /accounts/1/transfer", from.ID, to.ID, 200)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.InsufficientFunds); !ok {
		t.Fatalf("expected *apperrors.InsufficientFunds, got %T", err)
	}
}

func TestTransfer_SameAccount(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAccountService(db, NewIdempotencyService())

	account := &entities.Account{
		Name: "Self", Surname: "User", Email: "self@example.com",
		AddressLine1: "1 St", City: "London", Postcode: "E1", Country: "UK",
		Balance: 1000, Currency: "GBP",
	}
	account, _ = svc.Create("key-self", "POST /accounts", account)

	_, err := svc.Transfer("transfer-self", "POST /accounts/1/transfer", account.ID, account.ID, 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}

func TestTransfer_SourceNotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAccountService(db, NewIdempotencyService())

	to := &entities.Account{
		Name: "To", Surname: "User", Email: "to@example.com",
		AddressLine1: "2 St", City: "London", Postcode: "E2", Country: "UK",
		Balance: 500, Currency: "GBP",
	}
	to, _ = svc.Create("key-to", "POST /accounts", to)

	_, err := svc.Transfer("transfer-key", "POST /accounts/999/transfer", 999, to.ID, 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.AccountNotFound); !ok {
		t.Fatalf("expected *apperrors.AccountNotFound, got %T", err)
	}
}

func TestTransfer_DestinationNotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAccountService(db, NewIdempotencyService())

	from := &entities.Account{
		Name: "From", Surname: "User", Email: "from@example.com",
		AddressLine1: "1 St", City: "London", Postcode: "E1", Country: "UK",
		Balance: 1000, Currency: "GBP",
	}
	from, _ = svc.Create("key-from", "POST /accounts", from)

	_, err := svc.Transfer("transfer-key", "POST /accounts/1/transfer", from.ID, 999, 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.AccountNotFound); !ok {
		t.Fatalf("expected *apperrors.AccountNotFound, got %T", err)
	}
}

func TestTransfer_ConcurrentTransfersNoOverdraft(t *testing.T) {
	// Use shared-cache in-memory SQLite with busy timeout to handle concurrent access
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_busy_timeout=5000"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&entities.Account{}, &entities.IdempotencyRecord{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	// Limit to one open connection so SQLite serializes transactions properly
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)

	svc := NewAccountService(db, NewIdempotencyService())

	from := &entities.Account{
		Name: "From", Surname: "User", Email: "from@example.com",
		AddressLine1: "1 St", City: "London", Postcode: "E1", Country: "UK",
		Balance: 1000, Currency: "GBP",
	}
	to := &entities.Account{
		Name: "To", Surname: "User", Email: "to@example.com",
		AddressLine1: "2 St", City: "London", Postcode: "E2", Country: "UK",
		Balance: 0, Currency: "GBP",
	}

	from, _ = svc.Create("key-from", "POST /accounts", from)
	to, _ = svc.Create("key-to", "POST /accounts", to)

	// Attempt 10 concurrent transfers of 200 each from an account with 1000 balance.
	// With serialized access, exactly 5 should succeed and 5 should fail with InsufficientFunds.
	concurrency := 10
	var wg sync.WaitGroup
	results := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent-key-%d", index)
			_, err := svc.Transfer(key, "POST /accounts/1/transfer", from.ID, to.ID, 200)
			results <- err
		}(i)
	}

	wg.Wait()
	close(results)

	successCount := 0
	for err := range results {
		if err == nil {
			successCount++
		}
	}

	// Verify invariants: no overdraft and money is conserved
	finalFrom, _ := svc.FindByID(from.ID)
	finalTo, _ := svc.FindByID(to.ID)

	if finalFrom.Balance < 0 {
		t.Fatalf("overdraft detected: source balance is %d", finalFrom.Balance)
	}
	if finalFrom.Balance+finalTo.Balance != 1000 {
		t.Fatalf("money not conserved: source=%d + dest=%d != 1000", finalFrom.Balance, finalTo.Balance)
	}
	if successCount != 5 {
		t.Fatalf("expected 5 successful transfers, got %d", successCount)
	}
}

func TestTransfer_CurrencyMismatch(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAccountService(db, NewIdempotencyService())

	from := &entities.Account{
		Name: "From", Surname: "User", Email: "from@example.com",
		AddressLine1: "1 St", City: "London", Postcode: "E1", Country: "UK",
		Balance: 1000, Currency: "GBP",
	}
	to := &entities.Account{
		Name: "To", Surname: "User", Email: "to@example.com",
		AddressLine1: "2 St", City: "Berlin", Postcode: "10115", Country: "DE",
		Balance: 500, Currency: "EUR",
	}

	from, _ = svc.Create("key-gbp", "POST /accounts", from)
	to, _ = svc.Create("key-eur", "POST /accounts", to)

	_, err := svc.Transfer("transfer-cross", "POST /accounts/1/transfer", from.ID, to.ID, 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}
