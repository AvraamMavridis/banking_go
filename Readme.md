# Bank Account API

- Since one of the requirements was to run on any machine I decided to used docker, having it as the only prerequisite.
- I assumed that the API is for server-to-server communication, that's why I chose to use simple API keys for authentication, so `X-API-Key` is expected.
- I followed the pattern of idempotancy keys on headers to avoid duplicate side effects when a request is retried, so a `Idempotency-Key` is expected.
- Idempotency logic is extracted into its own `IdempotencyService` so it can be reused across different services without duplication.
- The Account entity accepts currency declaring the main currency of an account, but haven't implemented logic to convert amounts across currencies because that would require connecting to a service that provides live exchange rates.
- The amounts are store in cents, for example 1050 means 10.50 euro.
- In a production environment I would swap SQLite for PostgreSQL, but for the scope of this challenge SQLite keeps things simple and portable.
- I added a basic health endpoint that in a production like enviroment would had to connect to DB/cache etc and be used as Kubernetes liveness/readiness probe or any container orchestration health check
- Deposits and transfers use database transactions to ensure atomicity.
- Input validation is handled at the handler layer using `go-playground/validator`, while business rules (e.g. insufficient funds) are enforced in the service layer. I chose this package because is quite similar to Joi that I had used in the past and like for its simplicity.
- Deposits and transfers use database transactions to ensure atomicity.
- I am not super familiar with Go frameworks so I followed patterns that I would have used if I was developing this in HapiJS or Ruby on Rails.


## Prerequisites

- Docker

## Run with Docker

```bash
docker build -t bank-api .
docker run -p 8000:8000 -e API_KEY=your-secret-key bank-api
```

## Run locally

```bash
# Create a .env file from the example
cp .env.example .env
# Edit .env and set your API key

# Run
go run main.go
```

## Build

```bash
go build -o bank_api_go .
```

## Test

```bash
go test ./...
```

## API Endpoints

All endpoints require the `X-API-Key` header. Write endpoints also require an `Idempotency-Key` header (UUID v4).

### Health Check

```
GET /health
```

```bash
curl http://localhost:8000/health
```

### Get Account

```
GET /accounts/{id}
```

```bash
curl http://localhost:8000/accounts/1 \
  -H "X-API-Key: your-secret-key"
```

### Create Account

```
POST /accounts
```

```bash
curl -X POST http://localhost:8000/accounts \
  -H "X-API-Key: your-secret-key" \
  -H "Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John",
    "surname": "Doe",
    "email": "john@example.com",
    "addressLine1": "123 Main St",
    "city": "London",
    "postcode": "SW1A 1AA",
    "country": "UK",
    "balance": 1000,
    "currency": "GBP"
  }'
```

### Deposit

```
POST /accounts/{id}/deposit
```

```bash
curl -X POST http://localhost:8000/accounts/1/deposit \
  -H "X-API-Key: your-secret-key" \
  -H "Idempotency-Key: 660e8400-e29b-41d4-a716-446655440001" \
  -H "Content-Type: application/json" \
  -d '{"amount": 500}'
```

### Transfer

```
POST /accounts/{id}/transfer
```

```bash
curl -X POST http://localhost:8000/accounts/1/transfer \
  -H "X-API-Key: your-secret-key" \
  -H "Idempotency-Key: 770e8400-e29b-41d4-a716-446655440002" \
  -H "Content-Type: application/json" \
  -d '{"toAccountId": 2, "amount": 300}'
```
