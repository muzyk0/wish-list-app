package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	db "wish-list/internal/db/models"
	"wish-list/internal/repositories"

	"wish-list/internal/analytics"
	"wish-list/internal/auth"
	"wish-list/internal/aws"
	"wish-list/internal/cache"
	"wish-list/internal/config"
	"wish-list/internal/encryption"
	"wish-list/internal/middleware"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"

	"wish-list/internal/handlers"
	"wish-list/internal/services"
	"wish-list/internal/validation"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize Echo instance
	e := echo.New()

	// Set custom validator
	e.Validator = validation.NewValidator()

	// Set custom error handler
	e.HTTPErrorHandler = middleware.CustomHTTPErrorHandler

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	sqlxDB, err := db.New(dbCtx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func() {
		log.Println("Closing database connection...")
		if err := sqlxDB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Initialize JWT token manager
	tokenManager := auth.NewTokenManager(cfg.JWTSecret)

	// Initialize S3 client
	s3Client, err := aws.NewS3Client(cfg.AWSRegion, cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, cfg.AWSS3BucketName)
	if err != nil {
		log.Printf("Warning: Failed to initialize S3 client: %v", err)
		log.Println("Image upload functionality will be disabled")
	}

	// Initialize Redis cache
	var redisCache cache.CacheInterface
	redisCache, err = cache.NewRedisCache(
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.RedisDB,
		time.Duration(cfg.CacheTTLMinutes)*time.Minute,
	)
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis cache: %v", err)
		log.Println("Caching functionality will be disabled")
		redisCache = nil
	} else {
		defer func() {
			log.Println("Closing Redis connection...")
			if err := redisCache.(*cache.RedisCache).Close(); err != nil {
				log.Printf("Error closing Redis: %v", err)
			}
		}()
	}

	// Apply middleware in order
	e.Use(middleware.SecurityHeadersMiddleware())
	e.Use(middleware.RequestIDMiddleware())
	e.Use(middleware.LoggerMiddleware())
	e.Use(middleware.RecoverMiddleware())
	e.Use(middleware.CORSMiddleware(cfg.CorsAllowedOrigins))
	e.Use(middleware.TimeoutMiddleware(30 * time.Second))
	e.Use(middleware.RateLimiterMiddleware())

	// Initialize encryption service for PII protection (CR-004)
	var encryptionService *encryption.Service
	encryptionKey, encryptedKeyToStore, err := encryption.GetOrCreateDataKey(context.Background())
	if err != nil {
		// In production, encryption is required - fail fast
		if cfg.ServerEnv != "development" {
			log.Fatalf("FATAL: Failed to initialize encryption service in %s environment: %v. PII encryption is required.", cfg.ServerEnv, err)
		}
		log.Printf("Warning: Failed to initialize encryption service: %v. PII will not be encrypted.", err)
	} else {
		// If a new encrypted key was generated, instruct operator to persist it
		if encryptedKeyToStore != "" {
			fmt.Println("================================================================================")
			fmt.Println("IMPORTANT: A new KMS data encryption key has been generated.")
			fmt.Println("You MUST persist the following value to the ENCRYPTED_DATA_KEY environment")
			fmt.Println("variable or secret manager to prevent data loss on restart:")
			fmt.Println("")
			fmt.Printf("ENCRYPTED_DATA_KEY=%s\n", encryptedKeyToStore)
			fmt.Println("")
			fmt.Println("Without persisting this value, encrypted data will be unrecoverable.")
			fmt.Println("================================================================================")
		}

		encryptionService, err = encryption.NewService(encryptionKey)
		if err != nil {
			// In production, encryption is required - fail fast
			if cfg.ServerEnv != "development" {
				log.Fatalf("FATAL: Failed to create encryption service in %s environment: %v. PII encryption is required.", cfg.ServerEnv, err)
			}
			log.Printf("Warning: Failed to create encryption service: %v. PII will not be encrypted.", err)
		} else {
			log.Println("Encryption service initialized successfully for PII protection")
		}
	}

	// Initialize repositories
	var userRepo repositories.UserRepositoryInterface
	if encryptionService != nil {
		userRepo = repositories.NewUserRepositoryWithEncryption(sqlxDB, encryptionService)
	} else {
		userRepo = repositories.NewUserRepository(sqlxDB)
	}

	wishListRepo := repositories.NewWishListRepository(sqlxDB)
	giftItemRepo := repositories.NewGiftItemRepository(sqlxDB)
	templateRepo := repositories.NewTemplateRepository(sqlxDB)

	var reservationRepo repositories.ReservationRepositoryInterface
	if encryptionService != nil {
		reservationRepo = repositories.NewReservationRepositoryWithEncryption(sqlxDB, encryptionService)
	} else {
		reservationRepo = repositories.NewReservationRepository(sqlxDB)
	}

	// Initialize services
	analyticsService := analytics.NewAnalyticsService(cfg.AnalyticsEnabled)
	emailService := services.NewEmailService()
	userService := services.NewUserService(userRepo)
	wishListService := services.NewWishListService(wishListRepo, giftItemRepo, templateRepo, emailService, reservationRepo, redisCache)
	reservationService := services.NewReservationService(reservationRepo, giftItemRepo)
	accountCleanupService := services.NewAccountCleanupService(userRepo, wishListRepo, giftItemRepo, reservationRepo, emailService)

	// Initialize handlers with analytics integration
	userHandler := handlers.NewUserHandler(userService, tokenManager, accountCleanupService, analyticsService)
	wishListHandler := handlers.NewWishListHandler(wishListService)
	reservationHandler := handlers.NewReservationHandler(reservationService)

	// Start scheduled account cleanup job
	accountCleanupService.StartScheduledCleanup()

	// Initialize routes
	setupRoutes(e, userHandler, wishListHandler, reservationHandler, tokenManager, s3Client)

	// --- SERVER STARTUP AND SHUTDOWN ORCHESTRATION ---

	// Channel for server startup errors
	serverErrors := make(chan error, 1)

	port := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("🚀 Server is starting on port %s", port)

	// Run server in goroutines
	go func() {
		if err := e.Start(port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Channel for system signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Block until one of the events occurs
	select {
	case err := <-serverErrors:
		log.Fatalf("❌ Critical error, server failed to start: %v", err)

	case sig := <-stop:
		log.Printf("🚦 Received signal (%v), starting graceful shutdown...", sig)

		// Shutdown context (10 seconds timeout)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := e.Shutdown(shutdownCtx); err != nil {
			log.Printf("⚠️ Server forced to shutdown: %v", err)
			// If graceful shutdown failed, force close
			if err := e.Close(); err != nil {
				log.Printf("⚠️ Error closing server: %v", err)
			}
		}
	}

	log.Println("✅ Server stopped gracefully")
}

func setupRoutes(e *echo.Echo, userHandler *handlers.UserHandler, wishListHandler *handlers.WishListHandler, reservationHandler *handlers.ReservationHandler, tokenManager *auth.TokenManager, s3Client *aws.S3Client) {
	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	// User authentication endpoints
	authGroup := e.Group("/api/auth")
	authGroup.POST("/register", userHandler.Register)
	authGroup.POST("/login", userHandler.Login)

	// Example protected route using JWT
	protected := e.Group("/api/protected")
	protected.Use(auth.JWTMiddleware(tokenManager))
	protected.GET("/profile", userHandler.GetProfile)
	protected.PUT("/profile", userHandler.UpdateProfile)
	protected.DELETE("/account", userHandler.DeleteAccount)
	protected.GET("/export-data", userHandler.ExportUserData)

	// Example image upload endpoint (requires S3 client)
	if s3Client != nil {
		s3Handler := handlers.NewS3Handler(s3Client)

		imageUpload := e.Group("/api/images")
		imageUpload.Use(auth.JWTMiddleware(tokenManager))
		imageUpload.POST("/upload", s3Handler.UploadImage)
	}

	// Wish list endpoints
	wishListGroup := e.Group("/api/wishlists")
	wishListGroup.Use(auth.JWTMiddleware(tokenManager))
	wishListGroup.POST("", wishListHandler.CreateWishList)
	wishListGroup.GET("/:id", wishListHandler.GetWishList)
	wishListGroup.PUT("/:id", wishListHandler.UpdateWishList)
	wishListGroup.DELETE("/:id", wishListHandler.DeleteWishList)
	wishListGroup.GET("", wishListHandler.GetWishListsByOwner)

	// Gift item endpoints
	giftItemGroup := e.Group("/api/gift-items")
	giftItemGroup.Use(auth.JWTMiddleware(tokenManager))
	giftItemGroup.POST("/wishlist/:wishlistId", wishListHandler.CreateGiftItem)
	giftItemGroup.GET("/:id", wishListHandler.GetGiftItem)
	giftItemGroup.GET("/wishlist/:wishlistId", wishListHandler.GetGiftItemsByWishList)
	giftItemGroup.PUT("/:id", wishListHandler.UpdateGiftItem)
	giftItemGroup.DELETE("/:id", wishListHandler.DeleteGiftItem)
	giftItemGroup.POST("/:id/mark-purchased", wishListHandler.MarkGiftItemAsPurchased)

	// Public wish list endpoints
	publicListGroup := e.Group("/api/public/lists")
	publicListGroup.GET("/:slug", wishListHandler.GetWishListByPublicSlug)

	// Reservation endpoints
	reservationGroup := e.Group("/api/reservations")
	reservationGroup.Use(auth.JWTMiddleware(tokenManager))
	reservationGroup.POST("/wishlist/:wishlistId/item/:itemId", reservationHandler.CreateReservation)
	reservationGroup.DELETE("/wishlist/:wishlistId/item/:itemId", reservationHandler.CancelReservation)
	reservationGroup.GET("/user", reservationHandler.GetUserReservations)

	// Public reservation status endpoint
	publicReservationGroup := e.Group("/api/public/reservations")
	publicReservationGroup.GET("/list/:slug/item/:itemId", reservationHandler.GetReservationStatus)

	// Guest reservation endpoints
	guestReservationGroup := e.Group("/api/guest/reservations")
	guestReservationGroup.GET("", reservationHandler.GetGuestReservations)
}
