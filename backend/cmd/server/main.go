package main

import (
	"log"

	"wish-list/internal/app"
	"wish-list/internal/app/config"

	"github.com/joho/godotenv"
)

//	@title			Wish List API
//	@version		1.1
//	@description	A RESTful API for managing wish lists, gift items, and reservations.
//	@description	Features include user authentication, wish list management, gift item tracking, and reservation system.

//	@contact.name	API Support
//	@contact.email	support@wishlist.example.com

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:8080
//	@BasePath	/api

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
