package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bank_api_go/apperrors"
	"bank_api_go/entities"
	"bank_api_go/middleware"

	"github.com/go-playground/validator/v10"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http/response"
)

type AccountServicer interface {
	FindByID(id uint) (*entities.Account, error)
	Create(idempotencyKey string, fingerprint string, account *entities.Account) (*entities.Account, error)
	Deposit(idempotencyKey string, fingerprint string, id uint, amount int64) (*entities.Account, error)
	Transfer(idempotencyKey string, fingerprint string, fromID, toID uint, amount int64) (*entities.TransferResult, error)
}

type AccountHandler struct {
	service  AccountServicer
	validate *validator.Validate
}

func NewAccountHandler(service AccountServicer) *AccountHandler {
	return &AccountHandler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *AccountHandler) Register(app *gofr.App, authMiddleware func(http.Handler) http.Handler) {
	app.UseMiddleware(authMiddleware)

	app.GET("/accounts/{id}", h.GetByID)

	app.UseMiddleware(middleware.RequireIdempotencyKey)
	app.POST("/accounts", h.Create)
	app.POST("/accounts/{id}/deposit", h.Deposit)
	app.POST("/accounts/{id}/transfer", h.Transfer)
}

type CreateAccountRequest struct {
	Name         string  `json:"name" validate:"required,min=1,max=255"`
	Surname      string  `json:"surname" validate:"required,min=1,max=255"`
	Email        string  `json:"email" validate:"required,email,max=255"`
	Phone        *string `json:"phone" validate:"omitempty,max=50"`
	AddressLine1 string  `json:"addressLine1" validate:"required,min=1,max=255"`
	AddressLine2 *string `json:"addressLine2" validate:"omitempty,max=255"`
	City         string  `json:"city" validate:"required,min=1,max=100"`
	Postcode     string  `json:"postcode" validate:"required,min=1,max=20"`
	Country      string  `json:"country" validate:"required,min=1,max=100"`
	Balance      int64   `json:"balance" validate:"min=0"`
	Currency     string  `json:"currency" validate:"omitempty,len=3,uppercase"`
}

type DepositRequest struct {
	Amount int64 `json:"amount" validate:"required,gt=0"`
}

type TransferRequest struct {
	ToAccountID uint  `json:"toAccountId" validate:"required"`
	Amount      int64 `json:"amount" validate:"required,gt=0"`
}

func (h *AccountHandler) GetByID(ctx *gofr.Context) (any, error) {
	id, err := strconv.ParseUint(ctx.PathParam("id"), 10, 64)
	if err != nil {
		return nil, &apperrors.BadRequest{Message: "Invalid account ID"}
	}

	account, err := h.service.FindByID(uint(id))
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: account}, nil
}

func (h *AccountHandler) Create(ctx *gofr.Context) (any, error) {
	var req CreateAccountRequest
	if err := ctx.Bind(&req); err != nil {
		return nil, &apperrors.BadRequest{Message: "Invalid JSON body"}
	}

	if err := h.validate.Struct(req); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			return nil, &apperrors.BadRequest{Message: formatValidationErrors(ve)}
		}
		return nil, &apperrors.BadRequest{Message: "Validation failed"}
	}

	currency := req.Currency
	if currency == "" {
		currency = "EUR"
	}

	account := &entities.Account{
		Name:         req.Name,
		Surname:      req.Surname,
		Email:        req.Email,
		Phone:        req.Phone,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		Postcode:     req.Postcode,
		Country:      req.Country,
		Balance:      req.Balance,
		Currency:     currency,
	}

	idempotencyKey := getIdempotencyKey(ctx)
	saved, err := h.service.Create(idempotencyKey, "POST /accounts", account)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: saved}, nil
}

func (h *AccountHandler) Deposit(ctx *gofr.Context) (any, error) {
	id, err := strconv.ParseUint(ctx.PathParam("id"), 10, 64)
	if err != nil {
		return nil, &apperrors.BadRequest{Message: "Invalid account ID"}
	}

	var req DepositRequest
	if err := ctx.Bind(&req); err != nil {
		return nil, &apperrors.BadRequest{Message: "Invalid JSON body"}
	}

	if err := h.validate.Struct(req); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			return nil, &apperrors.BadRequest{Message: formatValidationErrors(ve)}
		}
		return nil, &apperrors.BadRequest{Message: "Validation failed"}
	}

	idempotencyKey := getIdempotencyKey(ctx)
	saved, err := h.service.Deposit(idempotencyKey, "POST /accounts/{id}/deposit", uint(id), req.Amount)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: saved}, nil
}

func (h *AccountHandler) Transfer(ctx *gofr.Context) (any, error) {
	id, err := strconv.ParseUint(ctx.PathParam("id"), 10, 64)
	if err != nil {
		return nil, &apperrors.BadRequest{Message: "Invalid account ID"}
	}

	var req TransferRequest
	if err := ctx.Bind(&req); err != nil {
		return nil, &apperrors.BadRequest{Message: "Invalid JSON body"}
	}

	if err := h.validate.Struct(req); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			return nil, &apperrors.BadRequest{Message: formatValidationErrors(ve)}
		}
		return nil, &apperrors.BadRequest{Message: "Validation failed"}
	}

	idempotencyKey := getIdempotencyKey(ctx)
	result, err := h.service.Transfer(idempotencyKey, "POST /accounts/{id}/transfer", uint(id), req.ToAccountID, req.Amount)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: result}, nil
}

func getIdempotencyKey(ctx *gofr.Context) string {
	if key, ok := ctx.Value(middleware.IdempotencyKeyCtx).(string); ok {
		return key
	}
	return ""
}

func formatValidationErrors(errs validator.ValidationErrors) string {
	messages := make([]string, len(errs))
	for i, e := range errs {
		messages[i] = e.Field() + " " + e.Tag()
	}
	data, err := json.Marshal(messages)
	if err != nil {
		return "Validation failed"
	}
	return string(data)
}
