package routes

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"bank_api_go/apperrors"
	"bank_api_go/entities"
	"bank_api_go/middleware"

	"github.com/gorilla/mux"
	"gofr.dev/pkg/gofr"
	gofrHTTP "gofr.dev/pkg/gofr/http"
	"gofr.dev/pkg/gofr/container"
	"gofr.dev/pkg/gofr/http/response"
	"gofr.dev/pkg/gofr/logging"
)

// mockAccountService implements AccountServicer for testing.
type mockAccountService struct {
	findByIDFn func(id uint) (*entities.Account, error)
	createFn   func(key, fp string, account *entities.Account) (*entities.Account, error)
	depositFn  func(key, fp string, id uint, amount int64) (*entities.Account, error)
	transferFn func(key, fp string, from, to uint, amount int64) (*entities.TransferResult, error)
}

func (mock *mockAccountService) FindByID(id uint) (*entities.Account, error) {
	if mock.findByIDFn != nil {
		return mock.findByIDFn(id)
	}
	return &entities.Account{ID: id, Name: "Test"}, nil
}

func (mock *mockAccountService) Create(key, fp string, account *entities.Account) (*entities.Account, error) {
	if mock.createFn != nil {
		return mock.createFn(key, fp, account)
	}
	account.ID = 1
	return account, nil
}

func (mock *mockAccountService) Deposit(key, fp string, id uint, amount int64) (*entities.Account, error) {
	if mock.depositFn != nil {
		return mock.depositFn(key, fp, id, amount)
	}
	return &entities.Account{ID: id, Balance: amount}, nil
}

func (mock *mockAccountService) Transfer(key, fp string, from, to uint, amount int64) (*entities.TransferResult, error) {
	if mock.transferFn != nil {
		return mock.transferFn(key, fp, from, to, amount)
	}
	return &entities.TransferResult{
		From: entities.Account{ID: from, Balance: 700},
		To:   entities.Account{ID: to, Balance: 800},
	}, nil
}

func newTestContext(method, path string, pathVars map[string]string, body []byte) *gofr.Context {
	httpReq := httptest.NewRequest(method, path, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	if pathVars != nil {
		httpReq = mux.SetURLVars(httpReq, pathVars)
	}

	// Set idempotency key in context
	ctx := context.WithValue(httpReq.Context(), middleware.IdempotencyKeyCtx, "test-idempotency-key")
	httpReq = httpReq.WithContext(ctx)

	req := gofrHTTP.NewRequest(httpReq)

	c := container.NewContainer(nil)
	c.Logger = logging.NewMockLogger(logging.ERROR)

	return &gofr.Context{
		Context:   httpReq.Context(),
		Request:   req,
		Container: c,
	}
}

// --- GetByID tests ---

func TestGetByID_Success(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	ctx := newTestContext("GET", "/accounts/1", map[string]string{"id": "1"}, nil)

	data, err := handler.GetByID(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, ok := data.(response.Raw)
	if !ok {
		t.Fatalf("expected response.Raw, got %T", data)
	}

	account, ok := raw.Data.(*entities.Account)
	if !ok {
		t.Fatalf("expected *entities.Account, got %T", raw.Data)
	}
	if account.ID != 1 {
		t.Fatalf("expected ID 1, got %d", account.ID)
	}
}

func TestGetByID_InvalidID(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	ctx := newTestContext("GET", "/accounts/abc", map[string]string{"id": "abc"}, nil)

	_, err := handler.GetByID(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := &mockAccountService{
		findByIDFn: func(id uint) (*entities.Account, error) {
			return nil, &apperrors.AccountNotFound{Message: "Account not found"}
		},
	}
	handler := NewAccountHandler(svc)
	ctx := newTestContext("GET", "/accounts/999", map[string]string{"id": "999"}, nil)

	_, err := handler.GetByID(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.AccountNotFound); !ok {
		t.Fatalf("expected *apperrors.AccountNotFound, got %T", err)
	}
}

// --- Create tests ---

func TestCreate_Success(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	body := []byte(`{
		"name": "John", "surname": "Doe", "email": "john@example.com",
		"addressLine1": "123 Main St", "city": "London", "postcode": "SW1A 1AA", "country": "UK"
	}`)
	ctx := newTestContext("POST", "/accounts", nil, body)

	data, err := handler.Create(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, ok := data.(response.Raw)
	if !ok {
		t.Fatalf("expected response.Raw, got %T", data)
	}

	account, ok := raw.Data.(*entities.Account)
	if !ok {
		t.Fatalf("expected *entities.Account, got %T", raw.Data)
	}
	if account.Name != "John" {
		t.Fatalf("expected name John, got %s", account.Name)
	}
	if account.Currency != "EUR" {
		t.Fatalf("expected default currency EUR, got %s", account.Currency)
	}
}

func TestCreate_ValidationError_MissingName(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	body := []byte(`{"surname": "Doe", "email": "john@example.com",
		"addressLine1": "123 Main St", "city": "London", "postcode": "SW1A 1AA", "country": "UK"}`)
	ctx := newTestContext("POST", "/accounts", nil, body)

	_, err := handler.Create(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}

func TestCreate_ValidationError_InvalidEmail(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	body := []byte(`{"name": "John", "surname": "Doe", "email": "not-an-email",
		"addressLine1": "123 Main St", "city": "London", "postcode": "SW1A 1AA", "country": "UK"}`)
	ctx := newTestContext("POST", "/accounts", nil, body)

	_, err := handler.Create(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	ctx := newTestContext("POST", "/accounts", nil, []byte(`{invalid`))

	_, err := handler.Create(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}

func TestCreate_CustomCurrency(t *testing.T) {
	var capturedCurrency string
	svc := &mockAccountService{
		createFn: func(key, fp string, account *entities.Account) (*entities.Account, error) {
			capturedCurrency = account.Currency
			account.ID = 1
			return account, nil
		},
	}
	handler := NewAccountHandler(svc)
	body := []byte(`{"name": "John", "surname": "Doe", "email": "john@example.com",
		"addressLine1": "123 Main St", "city": "London", "postcode": "SW1A 1AA", "country": "UK",
		"currency": "GBP"}`)
	ctx := newTestContext("POST", "/accounts", nil, body)

	_, err := handler.Create(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedCurrency != "GBP" {
		t.Fatalf("expected currency GBP, got %s", capturedCurrency)
	}
}

// --- Deposit tests ---

func TestDeposit_Success(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	body := []byte(`{"amount": 500}`)
	ctx := newTestContext("POST", "/accounts/1/deposit", map[string]string{"id": "1"}, body)

	data, err := handler.Deposit(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, ok := data.(response.Raw)
	if !ok {
		t.Fatalf("expected response.Raw, got %T", data)
	}
	if _, ok := raw.Data.(*entities.Account); !ok {
		t.Fatalf("expected *entities.Account, got %T", raw.Data)
	}
}

func TestDeposit_InvalidID(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	body := []byte(`{"amount": 500}`)
	ctx := newTestContext("POST", "/accounts/abc/deposit", map[string]string{"id": "abc"}, body)

	_, err := handler.Deposit(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}

func TestDeposit_InvalidAmount(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	body := []byte(`{"amount": 0}`)
	ctx := newTestContext("POST", "/accounts/1/deposit", map[string]string{"id": "1"}, body)

	_, err := handler.Deposit(ctx)
	if err == nil {
		t.Fatal("expected error for zero amount, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}

func TestDeposit_NegativeAmount(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	body := []byte(`{"amount": -100}`)
	ctx := newTestContext("POST", "/accounts/1/deposit", map[string]string{"id": "1"}, body)

	_, err := handler.Deposit(ctx)
	if err == nil {
		t.Fatal("expected error for negative amount, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}

func TestDeposit_InvalidJSON(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	ctx := newTestContext("POST", "/accounts/1/deposit", map[string]string{"id": "1"}, []byte(`{bad`))

	_, err := handler.Deposit(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}

func TestDeposit_ServiceError(t *testing.T) {
	svc := &mockAccountService{
		depositFn: func(key, fp string, id uint, amount int64) (*entities.Account, error) {
			return nil, &apperrors.AccountNotFound{Message: "Account not found"}
		},
	}
	handler := NewAccountHandler(svc)
	body := []byte(`{"amount": 500}`)
	ctx := newTestContext("POST", "/accounts/999/deposit", map[string]string{"id": "999"}, body)

	_, err := handler.Deposit(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.AccountNotFound); !ok {
		t.Fatalf("expected *apperrors.AccountNotFound, got %T", err)
	}
}

// --- Transfer tests ---

func TestTransfer_Success(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	body := []byte(`{"toAccountId": 2, "amount": 300}`)
	ctx := newTestContext("POST", "/accounts/1/transfer", map[string]string{"id": "1"}, body)

	data, err := handler.Transfer(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, ok := data.(response.Raw)
	if !ok {
		t.Fatalf("expected response.Raw, got %T", data)
	}
	result, ok := raw.Data.(*entities.TransferResult)
	if !ok {
		t.Fatalf("expected *entities.TransferResult, got %T", raw.Data)
	}
	if result.From.ID != 1 || result.To.ID != 2 {
		t.Fatalf("expected from=1 to=2, got from=%d to=%d", result.From.ID, result.To.ID)
	}
}

func TestTransfer_InvalidID(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	body := []byte(`{"toAccountId": 2, "amount": 300}`)
	ctx := newTestContext("POST", "/accounts/abc/transfer", map[string]string{"id": "abc"}, body)

	_, err := handler.Transfer(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}

func TestTransfer_InvalidJSON(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	ctx := newTestContext("POST", "/accounts/1/transfer", map[string]string{"id": "1"}, []byte(`{bad`))

	_, err := handler.Transfer(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}

func TestTransfer_MissingToAccountID(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	body := []byte(`{"amount": 300}`)
	ctx := newTestContext("POST", "/accounts/1/transfer", map[string]string{"id": "1"}, body)

	_, err := handler.Transfer(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}

func TestTransfer_InvalidAmount(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	body := []byte(`{"toAccountId": 2, "amount": 0}`)
	ctx := newTestContext("POST", "/accounts/1/transfer", map[string]string{"id": "1"}, body)

	_, err := handler.Transfer(ctx)
	if err == nil {
		t.Fatal("expected error for zero amount, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}

func TestTransfer_NegativeAmount(t *testing.T) {
	handler := NewAccountHandler(&mockAccountService{})
	body := []byte(`{"toAccountId": 2, "amount": -50}`)
	ctx := newTestContext("POST", "/accounts/1/transfer", map[string]string{"id": "1"}, body)

	_, err := handler.Transfer(ctx)
	if err == nil {
		t.Fatal("expected error for negative amount, got nil")
	}
	if _, ok := err.(*apperrors.BadRequest); !ok {
		t.Fatalf("expected *apperrors.BadRequest, got %T", err)
	}
}

func TestTransfer_ServiceError(t *testing.T) {
	svc := &mockAccountService{
		transferFn: func(key, fp string, from, to uint, amount int64) (*entities.TransferResult, error) {
			return nil, &apperrors.InsufficientFunds{}
		},
	}
	handler := NewAccountHandler(svc)
	body := []byte(`{"toAccountId": 2, "amount": 300}`)
	ctx := newTestContext("POST", "/accounts/1/transfer", map[string]string{"id": "1"}, body)

	_, err := handler.Transfer(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*apperrors.InsufficientFunds); !ok {
		t.Fatalf("expected *apperrors.InsufficientFunds, got %T", err)
	}
}

// --- Fingerprint tests ---

func TestCreate_PassesCorrectFingerprint(t *testing.T) {
	var capturedFingerprint string
	svc := &mockAccountService{
		createFn: func(key, fp string, account *entities.Account) (*entities.Account, error) {
			capturedFingerprint = fp
			account.ID = 1
			return account, nil
		},
	}
	handler := NewAccountHandler(svc)
	body := []byte(`{"name": "John", "surname": "Doe", "email": "john@example.com",
		"addressLine1": "123 Main St", "city": "London", "postcode": "SW1A 1AA", "country": "UK"}`)
	ctx := newTestContext("POST", "/accounts", nil, body)

	_, err := handler.Create(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFingerprint != "POST /accounts" {
		t.Fatalf("expected fingerprint 'POST /accounts', got %q", capturedFingerprint)
	}
}

func TestDeposit_PassesCorrectFingerprint(t *testing.T) {
	var capturedFingerprint string
	svc := &mockAccountService{
		depositFn: func(key, fp string, id uint, amount int64) (*entities.Account, error) {
			capturedFingerprint = fp
			return &entities.Account{ID: id, Balance: amount}, nil
		},
	}
	handler := NewAccountHandler(svc)
	body := []byte(`{"amount": 500}`)
	ctx := newTestContext("POST", "/accounts/1/deposit", map[string]string{"id": "1"}, body)

	_, err := handler.Deposit(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFingerprint != "POST /accounts/{id}/deposit" {
		t.Fatalf("expected fingerprint 'POST /accounts/{id}/deposit', got %q", capturedFingerprint)
	}
}

func TestTransfer_PassesCorrectFingerprint(t *testing.T) {
	var capturedFingerprint string
	svc := &mockAccountService{
		transferFn: func(key, fp string, from, to uint, amount int64) (*entities.TransferResult, error) {
			capturedFingerprint = fp
			return &entities.TransferResult{}, nil
		},
	}
	handler := NewAccountHandler(svc)
	body := []byte(`{"toAccountId": 2, "amount": 300}`)
	ctx := newTestContext("POST", "/accounts/1/transfer", map[string]string{"id": "1"}, body)

	_, err := handler.Transfer(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFingerprint != "POST /accounts/{id}/transfer" {
		t.Fatalf("expected fingerprint 'POST /accounts/{id}/transfer', got %q", capturedFingerprint)
	}
}
