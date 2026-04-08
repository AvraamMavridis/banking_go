package main

import (
	"log"
	"os"

	"bank_api_go/db"
	"github.com/joho/godotenv"
	"bank_api_go/middleware"
	"bank_api_go/routes"
	"bank_api_go/services"

	"gofr.dev/pkg/gofr"
)

func main() {
	_ = godotenv.Load()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "dev.sqlite3"
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is required")
	}

	database, err := db.Init(dbPath)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	idempotencyService := services.NewIdempotencyService()
	accountService := services.NewAccountService(database, idempotencyService)
	accountHandler := routes.NewAccountHandler(accountService)

	app := gofr.New()

	app.GET("/health", func(ctx *gofr.Context) (any, error) {
		return map[string]string{"status": "ok"}, nil
	})

	accountHandler.Register(app, middleware.RequireAPIKey(apiKey))
	app.Run()
}
