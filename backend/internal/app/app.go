package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wish-list/internal/app/swagger"

	"wish-list/internal/app/config"
	"wish-list/internal/app/database"
	"wish-list/internal/app/jobs"
	"wish-list/internal/app/server"

	authhttp "wish-list/internal/domain/auth/delivery/http"
	healthhttp "wish-list/internal/domain/health/delivery/http"
	itemhttp "wish-list/internal/domain/item/delivery/http"
	itemrepo "wish-list/internal/domain/item/repository"
	itemservice "wish-list/internal/domain/item/service"
	reservationhttp "wish-list/internal/domain/reservation/delivery/http"
	reservationrepo "wish-list/internal/domain/reservation/repository"
	reservationservice "wish-list/internal/domain/reservation/service"
	storagehttp "wish-list/internal/domain/storage/delivery/http"
	userhttp "wish-list/internal/domain/user/delivery/http"
	userrepo "wish-list/internal/domain/user/repository"
	userservice "wish-list/internal/domain/user/service"
	wishlisthttp "wish-list/internal/domain/wishlist/delivery/http"
	wishlistrepo "wish-list/internal/domain/wishlist/repository"
	wishlistservice "wish-list/internal/domain/wishlist/service"
	wishlistitemhttp "wish-list/internal/domain/wishlist_item/delivery/http"
	wishlistitemrepo "wish-list/internal/domain/wishlist_item/repository"
	wishlistitemservice "wish-list/internal/domain/wishlist_item/service"

	"wish-list/internal/pkg/analytics"
	"wish-list/internal/pkg/auth"
	"wish-list/internal/pkg/aws"
	"wish-list/internal/pkg/cache"
	"wish-list/internal/pkg/encryption"
	"wish-list/internal/pkg/logger"
	"wish-list/internal/pkg/validation"

	_ "wish-list/internal/app/swagger/docs" // Import generated Swagger docs
)

// App is the main application struct that wires all dependencies together.
type App struct {
	cfg    *config.Config
	db     *database.DB
	server *server.Server

	// Infrastructure
	tokenManager     *auth.TokenManager
	codeStore        *auth.CodeStore
	s3Client         *aws.S3Client
	redisCache       cache.CacheInterface
	encryptionSvc    *encryption.Service
	analyticsService *analytics.AnalyticsService

	// Background jobs
	accountCleanupService *jobs.AccountCleanupService

	// Domain handlers
	healthHandler       *healthhttp.Handler
	storageHandler      *storagehttp.Handler
	userHandler         *userhttp.Handler
	authHandler         *authhttp.Handler
	oauthHandler        *authhttp.OAuthHandler
	wishlistHandler     *wishlisthttp.Handler
	itemHandler         *itemhttp.Handler
	wishlistItemHandler *wishlistitemhttp.Handler
	reservationHandler  *reservationhttp.Handler
}

// New creates a new App instance, initializing all infrastructure, domain
// repositories, services, and handlers.
func New(cfg *config.Config) (*App, error) {
	// Initialize structured logger first
	logger.Initialize(cfg.ServerEnv)
	logger.Info("initializing application", "env", cfg.ServerEnv)

	a := &App{cfg: cfg}

	if err := a.initInfrastructure(); err != nil {
		return nil, fmt.Errorf("infrastructure init: %w", err)
	}

	a.initDomains()
	a.initServer()

	return a, nil
}

// initInfrastructure sets up database, encryption, cache, S3, token management.
func (a *App) initInfrastructure() error {
	// Database
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	db, err := database.New(dbCtx, a.cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database connection: %w", err)
	}
	a.db = db

	// JWT token manager
	a.tokenManager = auth.NewTokenManager(a.cfg.JWTSecret)

	// Code store for mobile handoff
	a.codeStore = auth.NewCodeStore()

	// S3 client (optional)
	s3Client, err := aws.NewS3Client(a.cfg.AWSRegion, a.cfg.AWSAccessKeyID, a.cfg.AWSSecretAccessKey, a.cfg.AWSS3BucketName)
	if err != nil {
		log.Printf("Warning: Failed to initialize S3 client: %v", err)
		log.Println("Image upload functionality will be disabled")
	}
	a.s3Client = s3Client

	// Redis cache (optional)
	redisCache, err := cache.NewRedisCache(
		a.cfg.RedisAddr,
		a.cfg.RedisPassword,
		a.cfg.RedisDB,
		time.Duration(a.cfg.CacheTTLMinutes)*time.Minute,
	)
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis cache: %v", err)
		log.Println("Caching functionality will be disabled")
	} else {
		a.redisCache = redisCache
	}

	// Encryption service for PII protection (CR-004)
	encryptionCtx, encryptionCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer encryptionCancel()

	encryptionKey, encryptedKeyToStore, err := encryption.GetOrCreateDataKey(encryptionCtx)
	if err != nil {
		if a.cfg.ServerEnv != "development" {
			return fmt.Errorf("encryption service required in %s: %w", a.cfg.ServerEnv, err)
		}
		log.Printf("Warning: Failed to initialize encryption service: %v. PII will not be encrypted.", err)
	} else {
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

		encSvc, err := encryption.NewService(encryptionKey)
		if err != nil {
			if a.cfg.ServerEnv != "development" {
				return fmt.Errorf("encryption service creation in %s: %w", a.cfg.ServerEnv, err)
			}
			log.Printf("Warning: Failed to create encryption service: %v. PII will not be encrypted.", err)
		} else {
			a.encryptionSvc = encSvc
			log.Println("Encryption service initialized successfully for PII protection")
		}
	}

	// Analytics
	a.analyticsService = analytics.NewAnalyticsService(a.cfg.AnalyticsEnabled)

	return nil
}

// initDomains wires all repositories, services, and handlers.
func (a *App) initDomains() {
	// --- Repositories ---

	var userRepo userrepo.UserRepositoryInterface
	if a.encryptionSvc != nil {
		userRepo = userrepo.NewUserRepositoryWithEncryption(a.db, a.encryptionSvc)
	} else {
		userRepo = userrepo.NewUserRepository(a.db)
	}

	wishlistRepo := wishlistrepo.NewWishListRepository(a.db)
	giftItemRepo := itemrepo.NewGiftItemRepository(a.db)
	giftItemReservationRepo := itemrepo.NewGiftItemReservationRepository(a.db)
	giftItemPurchaseRepo := itemrepo.NewGiftItemPurchaseRepository(a.db)
	wishlistItemRepo := wishlistitemrepo.NewWishlistItemRepository(a.db)

	var reservationRepo reservationrepo.ReservationRepositoryInterface
	if a.encryptionSvc != nil {
		reservationRepo = reservationrepo.NewReservationRepositoryWithEncryption(a.db, a.encryptionSvc)
	} else {
		reservationRepo = reservationrepo.NewReservationRepository(a.db)
	}

	// --- Services ---

	emailService := jobs.NewEmailService()
	userSvc := userservice.NewUserService(userRepo)
	wishlistSvc := wishlistservice.NewWishListService(wishlistRepo, giftItemRepo, giftItemReservationRepo, giftItemPurchaseRepo, emailService, reservationRepo, a.redisCache)
	itemSvc := itemservice.NewItemService(giftItemRepo, wishlistItemRepo)
	wishlistItemSvc := wishlistitemservice.NewWishlistItemService(wishlistRepo, giftItemRepo, wishlistItemRepo)
	reservationSvc := reservationservice.NewReservationService(reservationRepo, giftItemRepo, giftItemReservationRepo)
	a.accountCleanupService = jobs.NewAccountCleanupService(a.db, userRepo, wishlistRepo, giftItemRepo, reservationRepo, emailService)

	// --- Handlers ---

	a.healthHandler = healthhttp.NewHandler(a.db)
	a.userHandler = userhttp.NewHandler(userSvc, a.tokenManager, a.accountCleanupService, a.analyticsService)
	a.authHandler = authhttp.NewHandler(userSvc, a.tokenManager, a.codeStore)
	a.oauthHandler = authhttp.NewOAuthHandler(
		userRepo,
		a.tokenManager,
		a.cfg.GoogleClientID,
		a.cfg.GoogleClientSecret,
		a.cfg.FacebookClientID,
		a.cfg.FacebookClientSecret,
		a.cfg.OAuthRedirectURL,
		a.cfg.OAuthHTTPTimeout,
	)
	a.wishlistHandler = wishlisthttp.NewHandler(wishlistSvc)
	a.itemHandler = itemhttp.NewHandler(itemSvc)
	a.wishlistItemHandler = wishlistitemhttp.NewHandler(wishlistItemSvc)
	a.reservationHandler = reservationhttp.NewHandler(reservationSvc)

	if a.s3Client != nil {
		a.storageHandler = storagehttp.NewHandler(a.s3Client)
	}
}

// initServer creates the Echo server with middleware and registers all domain routes.
func (a *App) initServer() {
	a.server = server.New(a.cfg, validation.NewValidator())
	e := a.server.Echo

	// Swagger
	swagger.InitSwagger(e)

	// Auth middleware for protected routes
	authMiddleware := auth.JWTMiddleware(a.tokenManager)

	// Register all domain routes
	healthhttp.RegisterRoutes(e, a.healthHandler)
	userhttp.RegisterRoutes(e, a.userHandler, authMiddleware)
	authhttp.RegisterRoutes(e, a.authHandler, a.oauthHandler, authMiddleware)
	wishlisthttp.RegisterRoutes(e, a.wishlistHandler, authMiddleware)
	itemhttp.RegisterRoutes(e, a.itemHandler, authMiddleware)
	wishlistitemhttp.RegisterRoutes(e, a.wishlistItemHandler, authMiddleware)
	reservationhttp.RegisterRoutes(e, a.reservationHandler, authMiddleware)

	if a.storageHandler != nil {
		storagehttp.RegisterRoutes(e, a.storageHandler, a.tokenManager)
	}
}

// Run starts the application: background jobs and HTTP server.
// Blocks until a shutdown signal is received.
func (a *App) Run() error {
	// Application context for lifecycle management
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// Start code store cleanup goroutine
	a.codeStore.StartCleanupRoutine(appCtx)

	// Start background jobs
	a.accountCleanupService.StartScheduledCleanup(appCtx)

	// Start HTTP server
	port := fmt.Sprintf(":%d", a.cfg.ServerPort)
	log.Printf("Server is starting on port %s", port)

	serverErrors := make(chan error, 1)
	go func() {
		if err := a.server.Echo.Start(port); err != nil {
			serverErrors <- err
		}
	}()

	// Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server failed to start: %w", err)

	case sig := <-stop:
		log.Printf("Received signal (%v), starting graceful shutdown...", sig)
		appCancel()
		return a.Shutdown(context.Background())
	}
}

// Shutdown gracefully shuts down the application.
func (a *App) Shutdown(ctx context.Context) error {
	log.Println("Stopping background services...")

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	if err := a.server.Echo.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		if closeErr := a.server.Echo.Close(); closeErr != nil {
			log.Printf("Error closing server: %v", closeErr)
		}
	}

	// Close Redis
	if a.redisCache != nil {
		log.Println("Closing Redis connection...")
		if err := a.redisCache.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		}
	}

	// Close database
	log.Println("Closing database connection...")
	if err := a.db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server stopped gracefully")
	return nil
}
