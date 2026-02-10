## 📁 Пример структуры проекта

```
wishlist-app/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── app/
│   │   ├── config/
│   │   │   ├── config.go           # Конфигурация через godotenv
│   │   │   └── validator.go        # Инициализация валидатора
│   │   ├── database/
│   │   │   ├── postgres.go         # Подключение через pgx/sqlx
│   │   │   └── migrations.go       # Управление миграциями
│   │   ├── server/
│   │   │   ├── server.go
│   │   │   └── router.go           # Маршрутизация со swagger
│   │   ├── swagger/
│   │   │   └── docs.go             # Инициализация swagger
│   │   └── app.go
│   ├── pkg/
│   │   ├── auth/
│   │   │   ├── jwt/
│   │   │   │   └── manager.go
│   │   │   └── middleware/
│   │   │       └── auth.go
│   │   ├── logger/
│   │   │   └── logger.go
│   │   ├── response/
│   │   │   └── response.go
│   │   └── validator/
│   │       └── custom_validator.go # Кастомные правила валидации
│   └── domain/
│       ├── auth/
│       │   ├── delivery/
│       │   │   └── http/
│       │   │       ├── handler.go  # С аннотациями swaggo
│       │   │       ├── dto/
│       │   │       │   ├── requests.go
│       │   │       │   └── responses.go
│       │   │       └── routes.go
│       │   ├── service/
│       │   │   └── auth_service.go
│       │   ├── repository/
│       │   │   └── auth_repository.go
│       │   └── models/
│       │       └── auth.go
│       ├── user/                   # Аналогичная структура
│       ├── wishlist/
│       └── item/
├── api/
│   ├── docs/                       # Генерируемая документация
│   │   ├── docs.go
│   │   ├── swagger.json
│   │   └── swagger.yaml
│   └── openapi/                    # Ручные спецификации
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile
│   │   ├── Dockerfile.dev
│   │   └── .dockerignore
│   └── docker-compose.yml
├── migrations/                     # SQL миграции
│   ├── 001_create_users.up.sql
│   ├── 001_create_users.down.sql
│   └── ...
├── test/
│   ├── integration/
│   │   └── auth_test.go
│   ├── mocks/                      # Генерируемые моки
│   └── testdata/
├── scripts/
│   ├── migrate.sh
│   └── generate-mocks.sh
├── .env.example
├── .env.local
├── .air.toml                      # Hot reload конфиг
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 📝 Детальное описание файлов с использованием указанных библиотек

### 1. **Конфигурация через godotenv** (`internal/app/config/config.go`)

```go
package config

import (
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"
    "time"
    
    "github.com/joho/godotenv"
)

type Config struct {
    App      AppConfig
    Server   ServerConfig
    Database DatabaseConfig
    Auth     AuthConfig
    Log      LogConfig
    Swagger  SwaggerConfig
}

type AppConfig struct {
    Name        string `mapstructure:"APP_NAME"`
    Version     string `mapstructure:"APP_VERSION"`
    Environment string `mapstructure:"APP_ENV"`
    Debug       bool   `mapstructure:"APP_DEBUG"`
}

type ServerConfig struct {
    Host         string        `mapstructure:"HOST"`
    Port         string        `mapstructure:"PORT"`
    ReadTimeout  time.Duration `mapstructure:"SERVER_READ_TIMEOUT"`
    WriteTimeout time.Duration `mapstructure:"SERVER_WRITE_TIMEOUT"`
    IdleTimeout  time.Duration `mapstructure:"SERVER_IDLE_TIMEOUT"`
}

type DatabaseConfig struct {
    Host            string        `mapstructure:"DB_HOST"`
    Port            string        `mapstructure:"DB_PORT"`
    User            string        `mapstructure:"DB_USER"`
    Password        string        `mapstructure:"DB_PASSWORD"`
    Name            string        `mapstructure:"DB_NAME"`
    SSLMode         string        `mapstructure:"DB_SSL_MODE"`
    MaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS"`
    MaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS"`
    ConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`
    MigrationsPath  string        `mapstructure:"DB_MIGRATIONS_PATH"`
}

type AuthConfig struct {
    JWTSecret          string        `mapstructure:"JWT_SECRET"`
    AccessTokenTTL     time.Duration `mapstructure:"JWT_ACCESS_TOKEN_TTL"`
    RefreshTokenTTL    time.Duration `mapstructure:"JWT_REFRESH_TOKEN_TTL"`
    Issuer             string        `mapstructure:"JWT_ISSUER"`
    Audience           string        `mapstructure:"JWT_AUDIENCE"`
}

type LogConfig struct {
    Level  string `mapstructure:"LOG_LEVEL"`
    Format string `mapstructure:"LOG_FORMAT"`
    Output string `mapstructure:"LOG_OUTPUT"`
}

type SwaggerConfig struct {
    Enabled bool   `mapstructure:"SWAGGER_ENABLED"`
    Host    string `mapstructure:"SWAGGER_HOST"`
}

// Load загружает конфигурацию из .env файла и переменных окружения
func Load() (*Config, error) {
    // 1. Загрузка .env файла (опционально)
    envFile := ".env"
    if env := os.Getenv("APP_ENV"); env != "" {
        envFile = fmt.Sprintf(".env.%s", env)
    }
    
    // Пытаемся загрузить .env файл, но не падаем, если его нет
    if err := godotenv.Load(envFile); err != nil {
        log.Printf("Note: No %s file found, using environment variables only", envFile)
    }
    
    // 2. Загрузка значений в структуру
    cfg := &Config{
        App: AppConfig{
            Name:        getEnv("APP_NAME", "wishlist-app"),
            Version:     getEnv("APP_VERSION", "1.0.0"),
            Environment: getEnv("APP_ENV", "development"),
            Debug:       getEnvAsBool("APP_DEBUG", true),
        },
        Server: ServerConfig{
            Host:         getEnv("HOST", "0.0.0.0"),
            Port:         getEnv("PORT", "8080"),
            ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 30*time.Second),
            WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
            IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
        },
        Database: DatabaseConfig{
            Host:            getEnv("DB_HOST", "localhost"),
            Port:            getEnv("DB_PORT", "5432"),
            User:            getEnv("DB_USER", "postgres"),
            Password:        getEnv("DB_PASSWORD", "postgres"),
            Name:            getEnv("DB_NAME", "wishlist_db"),
            SSLMode:         getEnv("DB_SSL_MODE", "disable"),
            MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
            MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
            ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
            MigrationsPath:  getEnv("DB_MIGRATIONS_PATH", "./migrations"),
        },
        Auth: AuthConfig{
            JWTSecret:          getEnv("JWT_SECRET", ""),
            AccessTokenTTL:     getEnvAsDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
            RefreshTokenTTL:    getEnvAsDuration("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour),
            Issuer:             getEnv("JWT_ISSUER", "wishlist-app"),
            Audience:           getEnv("JWT_AUDIENCE", "wishlist-app"),
        },
        Log: LogConfig{
            Level:  getEnv("LOG_LEVEL", "info"),
            Format: getEnv("LOG_FORMAT", "json"),
            Output: getEnv("LOG_OUTPUT", "stdout"),
        },
        Swagger: SwaggerConfig{
            Enabled: getEnvAsBool("SWAGGER_ENABLED", true),
            Host:    getEnv("SWAGGER_HOST", "localhost:8080"),
        },
    }
    
    // 3. Валидация обязательных полей
    if err := validateConfig(cfg); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }
    
    return cfg, nil
}

// validateConfig валидирует конфигурацию
func validateConfig(cfg *Config) error {
    if cfg.Auth.JWTSecret == "" {
        return fmt.Errorf("JWT_SECRET is required")
    }
    
    if cfg.Database.Host == "" {
        return fmt.Errorf("DB_HOST is required")
    }
    
    return nil
}

// Вспомогательные функции для работы с переменными окружения
func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    valueStr := getEnv(key, "")
    if value, err := strconv.Atoi(valueStr); err == nil {
        return value
    }
    return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
    valueStr := getEnv(key, "")
    if value, err := strconv.ParseBool(valueStr); err == nil {
        return value
    }
    return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
    valueStr := getEnv(key, "")
    if value, err := time.ParseDuration(valueStr); err == nil {
        return value
    }
    return defaultValue
}

func getEnvAsSlice(key string, separator string, defaultValue []string) []string {
    valueStr := getEnv(key, "")
    if valueStr == "" {
        return defaultValue
    }
    return strings.Split(valueStr, separator)
}
```

### 2. **Подключение к БД через pgx/sqlx** (`internal/app/database/postgres.go`)

```go
package database

import (
    "context"
    "fmt"
    "time"
    
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/jackc/pgx/v5/stdlib"
    "github.com/jmoiron/sqlx"
    "wishlist-app/internal/app/config"
)

// DB обертка над sqlx.DB и pgxpool.Pool
type DB struct {
    sqlxDB *sqlx.DB
    pgxPool *pgxpool.Pool
}

// NewPostgres создает новое подключение к PostgreSQL
func NewPostgres(cfg config.DatabaseConfig) (*DB, error) {
    // Формируем DSN строку
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
    )
    
    // 1. Подключение через sqlx (для совместимости)
    sqlxDB, err := connectWithSqlx(dsn, cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to connect with sqlx: %w", err)
    }
    
    // 2. Подключение через pgxpool (для лучшей производительности)
    pgxPool, err := connectWithPgxPool(dsn, cfg)
    if err != nil {
        sqlxDB.Close()
        return nil, fmt.Errorf("failed to connect with pgxpool: %w", err)
    }
    
    return &DB{
        sqlxDB:  sqlxDB,
        pgxPool: pgxPool,
    }, nil
}

// connectWithSqlx подключение через sqlx с драйвером pgx
func connectWithSqlx(dsn string, cfg config.DatabaseConfig) (*sqlx.DB, error) {
    // Создаем конфигурацию pgx
    pgxConfig, err := pgx.ParseConfig(dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to parse pgx config: %w", err)
    }
    
    // Конвертируем pgx.Config в sql.DB через stdlib
    connStr := stdlib.RegisterConnConfig(pgxConfig)
    
    // Создаем sqlx.DB
    db, err := sqlx.Open("pgx", connStr)
    if err != nil {
        return nil, fmt.Errorf("failed to open sqlx database: %w", err)
    }
    
    // Настраиваем пул соединений
    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    
    // Проверяем подключение
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    return db, nil
}

// connectWithPgxPool подключение через pgxpool
func connectWithPgxPool(dsn string, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
    // Создаем конфигурацию пула
    poolConfig, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to parse pool config: %w", err)
    }
    
    // Настраиваем пул
    poolConfig.MaxConns = int32(cfg.MaxOpenConns)
    poolConfig.MinConns = int32(cfg.MaxIdleConns)
    poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
    poolConfig.MaxConnIdleTime = 30 * time.Minute
    
    // Подключаемся
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create connection pool: %w", err)
    }
    
    // Проверяем подключение
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping pool: %w", err)
    }
    
    return pool, nil
}

// GetSqlx возвращает sqlx.DB
func (db *DB) GetSqlx() *sqlx.DB {
    return db.sqlxDB
}

// GetPgxPool возвращает pgxpool.Pool
func (db *DB) GetPgxPool() *pgxpool.Pool {
    return db.pgxPool
}

// Close закрывает все соединения
func (db *DB) Close() {
    if db.sqlxDB != nil {
        db.sqlxDB.Close()
    }
    if db.pgxPool != nil {
        db.pgxPool.Close()
    }
}

// HealthCheck проверяет доступность БД
func (db *DB) HealthCheck(ctx context.Context) error {
    if err := db.sqlxDB.PingContext(ctx); err != nil {
        return fmt.Errorf("sqlx ping failed: %w", err)
    }
    
    if err := db.pgxPool.Ping(ctx); err != nil {
        return fmt.Errorf("pgxpool ping failed: %w", err)
    }
    
    return nil
}

// BeginTx начинает транзакцию с использованием pgx
func (db *DB) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
    return db.pgxPool.BeginTx(ctx, opts)
}
```

### 3. **Управление миграциями** (`internal/app/database/migrations.go`)

```go
package database

import (
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    "github.com/golang-migrate/migrate/v4/source/iofs"
    "wishlist-app/internal/app/config"
)

// Migrator управляет миграциями БД
type Migrator struct {
    migrate *migrate.Migrate
    cfg     config.DatabaseConfig
}

// NewMigrator создает новый мигратор
func NewMigrator(cfg config.DatabaseConfig) (*Migrator, error) {
    // Формируем DSN для миграций
    dsn := fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=%s",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode,
    )
    
    // Проверяем существование директории с миграциями
    if _, err := os.Stat(cfg.MigrationsPath); os.IsNotExist(err) {
        return nil, fmt.Errorf("migrations directory does not exist: %s", cfg.MigrationsPath)
    }
    
    // Создаем абсолютный путь
    absPath, err := filepath.Abs(cfg.MigrationsPath)
    if err != nil {
        return nil, fmt.Errorf("failed to get absolute path: %w", err)
    }
    
    // Создаем source URL для файловой системы
    sourceURL := fmt.Sprintf("file://%s", absPath)
    
    // Создаем мигратор
    m, err := migrate.New(sourceURL, dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to create migrate instance: %w", err)
    }
    
    return &Migrator{
        migrate: m,
        cfg:     cfg,
    }, nil
}

// Up применяет все миграции
func (m *Migrator) Up() error {
    if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("failed to apply migrations: %w", err)
    }
    return nil
}

// Down откатывает все миграции
func (m *Migrator) Down() error {
    if err := m.migrate.Down(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("failed to rollback migrations: %w", err)
    }
    return nil
}

// Steps применяет или откатывает указанное количество миграций
func (m *Migrator) Steps(n int) error {
    if err := m.migrate.Steps(n); err != nil {
        return fmt.Errorf("failed to apply %d steps: %w", n, err)
    }
    return nil
}

// Version возвращает текущую версию миграции
func (m *Migrator) Version() (uint, bool, error) {
    version, dirty, err := m.migrate.Version()
    if err != nil {
        return 0, false, fmt.Errorf("failed to get migration version: %w", err)
    }
    return version, dirty, nil
}

// Force устанавливает указанную версию миграции
func (m *Migrator) Force(version int) error {
    if err := m.migrate.Force(version); err != nil {
        return fmt.Errorf("failed to force migration version: %w", err)
    }
    return nil
}

// Drop удаляет все таблицы (только для тестов!)
func (m *Migrator) Drop() error {
    if err := m.migrate.Drop(); err != nil {
        return fmt.Errorf("failed to drop database: %w", err)
    }
    return nil
}

// Close закрывает мигратор
func (m *Migrator) Close() error {
    if sourceErr, dbErr := m.migrate.Close(); sourceErr != nil || dbErr != nil {
        return fmt.Errorf("source error: %v, database error: %v", sourceErr, dbErr)
    }
    return nil
}
```

### 4. **Инициализация валидатора** (`internal/app/config/validator.go`)

```go
package config

import (
    "github.com/go-playground/validator/v10"
    "github.com/labstack/echo/v4"
)

// CustomValidator кастомный валидатор для Echo
type CustomValidator struct {
    validator *validator.Validate
}

// NewCustomValidator создает новый валидатор с кастомными правилами
func NewCustomValidator() *CustomValidator {
    v := validator.New()
    
    // Регистрируем кастомные валидации
    registerCustomValidations(v)
    
    return &CustomValidator{validator: v}
}

// Validate реализует интерфейс echo.Validator
func (cv *CustomValidator) Validate(i interface{}) error {
    return cv.validator.Struct(i)
}

// registerCustomValidations регистрирует кастомные правила валидации
func registerCustomValidations(v *validator.Validate) {
    // Валидация пароля
    v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
        password := fl.Field().String()
        if len(password) < 8 {
            return false
        }
        
        // Проверяем наличие цифр, букв в верхнем и нижнем регистре
        hasNumber := false
        hasUpper := false
        hasLower := false
        
        for _, char := range password {
            switch {
            case '0' <= char && char <= '9':
                hasNumber = true
            case 'A' <= char && char <= 'Z':
                hasUpper = true
            case 'a' <= char && char <= 'z':
                hasLower = true
            }
        }
        
        return hasNumber && hasUpper && hasLower
    })
    
    // Валидация цены
    v.RegisterValidation("price", func(fl validator.FieldLevel) bool {
        price := fl.Field().Float()
        return price >= 0
    })
    
    // Валидация URL
    v.RegisterValidation("url", func(fl validator.FieldLevel) bool {
        url := fl.Field().String()
        if url == "" {
            return true // Пустой URL допустим
        }
        return validator.New().Var(url, "url") == nil
    })
    
    // Валидация массива тегов
    v.RegisterValidation("tags", func(fl validator.FieldLevel) bool {
        tags := fl.Field().Interface().([]string)
        if len(tags) > 10 {
            return false // Максимум 10 тегов
        }
        
        for _, tag := range tags {
            if len(tag) > 50 || len(tag) < 1 {
                return false
            }
        }
        
        return true
    })
}

// InitValidator инициализирует валидатор в Echo
func InitValidator(e *echo.Echo) {
    e.Validator = NewCustomValidator()
}

// ValidationErrors преобразует ошибки валидации в map
func ValidationErrors(err error) map[string]string {
    errs := make(map[string]string)
    
    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        for _, e := range validationErrors {
            field := e.Field()
            tag := e.Tag()
            
            // Пользовательские сообщения об ошибках
            switch tag {
            case "required":
                errs[field] = "Это поле обязательно для заполнения"
            case "email":
                errs[field] = "Неверный формат email"
            case "password":
                errs[field] = "Пароль должен содержать минимум 8 символов, цифры и буквы в верхнем и нижнем регистре"
            case "min":
                errs[field] = fmt.Sprintf("Минимальная длина: %s символов", e.Param())
            case "max":
                errs[field] = fmt.Sprintf("Максимальная длина: %s символов", e.Param())
            case "url":
                errs[field] = "Неверный формат URL"
            case "price":
                errs[field] = "Цена не может быть отрицательной"
            case "tags":
                errs[field] = "Максимум 10 тегов, каждый до 50 символов"
            default:
                errs[field] = fmt.Sprintf("Ошибка валидации: %s", tag)
            }
        }
    }
    
    return errs
}
```

### 5. **Обработчик с аннотациями Swaggo** (`internal/domain/auth/delivery/http/handler.go`)

```go
package handler

import (
    "net/http"
    "time"
    
    "github.com/labstack/echo/v4"
    "wishlist-app/internal/domain/auth/delivery/http/dto"
    "wishlist-app/internal/domain/auth/service"
    "wishlist-app/internal/pkg/response"
)

// AuthHandler обработчик аутентификации
// @title Wishlist App API
// @version 1.0
// @description API для управления вишлистами
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@wishlistapp.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token
type AuthHandler struct {
    authService service.AuthService
}

// NewAuthHandler создает новый обработчик
func NewAuthHandler(authService service.AuthService) *AuthHandler {
    return &AuthHandler{
        authService: authService,
    }
}

// Register регистрирует нового пользователя
// @Summary Регистрация пользователя
// @Description Создает нового пользователя в системе
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
    var req dto.RegisterRequest
    
    if err := c.Bind(&req); err != nil {
        return response.BadRequest(c, "Неверный формат запроса")
    }
    
    if err := c.Validate(req); err != nil {
        return response.ValidationError(c, err)
    }
    
    authResponse, err := h.authService.Register(c.Request().Context(), req.ToDomain())
    if err != nil {
        return h.handleServiceError(c, err)
    }
    
    resp := dto.NewAuthResponse(authResponse)
    return response.Created(c, "Пользователь успешно зарегистрирован", resp)
}

// Login аутентифицирует пользователя
// @Summary Вход в систему
// @Description Аутентифицирует пользователя и возвращает токены
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Данные для входа"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
    var req dto.LoginRequest
    
    if err := c.Bind(&req); err != nil {
        return response.BadRequest(c, "Неверный формат запроса")
    }
    
    if err := c.Validate(req); err != nil {
        return response.ValidationError(c, err)
    }
    
    authResponse, err := h.authService.Login(c.Request().Context(), req.ToDomain())
    if err != nil {
        return h.handleServiceError(c, err)
    }
    
    resp := dto.NewAuthResponse(authResponse)
    
    // Устанавливаем refresh token в http-only cookie
    h.setRefreshTokenCookie(c, authResponse.RefreshToken)
    
    return response.Success(c, "Вход выполнен успешно", resp)
}

// Refresh обновляет токены
// @Summary Обновление токенов
// @Description Обновляет access token с помощью refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "Refresh token"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c echo.Context) error {
    var req dto.RefreshRequest
    
    // Пытаемся получить refresh token из cookie
    refreshToken := h.getRefreshTokenFromCookie(c)
    if refreshToken == "" {
        // Если нет в cookie, пробуем из тела запроса
        if err := c.Bind(&req); err != nil {
            return response.BadRequest(c, "Refresh token не предоставлен")
        }
        refreshToken = req.RefreshToken
    }
    
    if refreshToken == "" {
        return response.Unauthorized(c, "Refresh token обязателен")
    }
    
    authResponse, err := h.authService.RefreshTokens(c.Request().Context(), refreshToken)
    if err != nil {
        return h.handleServiceError(c, err)
    }
    
    resp := dto.NewAuthResponse(authResponse)
    
    // Обновляем refresh token в cookie
    h.setRefreshTokenCookie(c, authResponse.RefreshToken)
    
    return response.Success(c, "Токены успешно обновлены", resp)
}

// Logout выходит из системы
// @Summary Выход из системы
// @Description Завершает сессию пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
    userID := c.Get("userID").(string)
    refreshToken := h.getRefreshTokenFromCookie(c)
    
    if err := h.authService.Logout(c.Request().Context(), userID, refreshToken); err != nil {
        return h.handleServiceError(c, err)
    }
    
    h.clearRefreshTokenCookie(c)
    return response.Success(c, "Выход выполнен успешно", nil)
}

// GetProfile возвращает профиль пользователя
// @Summary Получение профиля
// @Description Возвращает информацию о текущем пользователе
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ProfileResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) GetProfile(c echo.Context) error {
    userID := c.Get("userID").(string)
    
    profile, err := h.authService.GetProfile(c.Request().Context(), userID)
    if err != nil {
        return h.handleServiceError(c, err)
    }
    
    resp := dto.NewProfileResponse(profile)
    return response.Success(c, "Профиль получен успешно", resp)
}

// handleServiceError обрабатывает ошибки сервиса
func (h *AuthHandler) handleServiceError(c echo.Context, err error) error {
    switch err {
    case service.ErrUserAlreadyExists:
        return response.Conflict(c, "Пользователь с таким email уже существует")
    case service.ErrInvalidCredentials:
        return response.Unauthorized(c, "Неверные учетные данные")
    case service.ErrUserNotFound:
        return response.NotFound(c, "Пользователь не найден")
    case service.ErrInvalidToken:
        return response.Unauthorized(c, "Неверный токен")
    case service.ErrTokenExpired:
        return response.Unauthorized(c, "Срок действия токена истек")
    default:
        return response.InternalServerError(c, "Внутренняя ошибка сервера")
    }
}

// setRefreshTokenCookie устанавливает refresh token в cookie
func (h *AuthHandler) setRefreshTokenCookie(c echo.Context, token string) {
    cookie := new(http.Cookie)
    cookie.Name = "refresh_token"
    cookie.Value = token
    cookie.Path = "/"
    cookie.HttpOnly = true
    cookie.Secure = false // В production должно быть true
    cookie.SameSite = http.SameSiteStrictMode
    cookie.MaxAge = int(7 * 24 * time.Hour / time.Second) // 7 дней
    
    c.SetCookie(cookie)
}

// getRefreshTokenFromCookie получает refresh token из cookie
func (h *AuthHandler) getRefreshTokenFromCookie(c echo.Context) string {
    cookie, err := c.Cookie("refresh_token")
    if err != nil {
        return ""
    }
    return cookie.Value
}

// clearRefreshTokenCookie очищает refresh token cookie
func (h *AuthHandler) clearRefreshTokenCookie(c echo.Context) {
    cookie := new(http.Cookie)
    cookie.Name = "refresh_token"
    cookie.Value = ""
    cookie.Path = "/"
    cookie.HttpOnly = true
    cookie.Secure = false
    cookie.SameSite = http.SameSiteStrictMode
    cookie.MaxAge = -1
    
    c.SetCookie(cookie)
}
```

### 6. **DTO с валидацией** (`internal/domain/auth/delivery/http/dto/requests.go`)

```go
package dto

import (
    "wishlist-app/internal/domain/auth/models"
)

// RegisterRequest запрос на регистрацию
type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,password,min=8"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
}

// ToDomain преобразует DTO в доменную модель
func (r *RegisterRequest) ToDomain() *models.User {
    return &models.User{
        Email:    r.Email,
        Password: r.Password,
        Name:     r.Name,
    }
}

// LoginRequest запрос на вход
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

// ToDomain преобразует DTO в доменную модель
func (r *LoginRequest) ToDomain() (string, string) {
    return r.Email, r.Password
}

// RefreshRequest запрос на обновление токена
type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" validate:"required"`
}

// UpdateProfileRequest запрос на обновление профиля
type UpdateProfileRequest struct {
    Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
    Email *string `json:"email,omitempty" validate:"omitempty,email"`
}

// ChangePasswordRequest запрос на смену пароля
type ChangePasswordRequest struct {
    OldPassword string `json:"old_password" validate:"required"`
    NewPassword string `json:"new_password" validate:"required,password,min=8"`
}

// ResetPasswordRequest запрос на сброс пароля
type ResetPasswordRequest struct {
    Email string `json:"email" validate:"required,email"`
}

// ConfirmResetPasswordRequest запрос на подтверждение сброса пароля
type ConfirmResetPasswordRequest struct {
    Token    string `json:"token" validate:"required"`
    Password string `json:"password" validate:"required,password,min=8"`
}
```

### 7. **Swagger инициализация** (`internal/app/swagger/docs.go`)

```go
package swagger

import (
    "fmt"
    "github.com/labstack/echo/v4"
    echoSwagger "github.com/swaggo/echo-swagger"
    "wishlist-app/internal/app/config"
)

// InitSwagger инициализирует Swagger документацию
func InitSwagger(e *echo.Echo, cfg config.SwaggerConfig) {
    if !cfg.Enabled {
        return
    }
    
    // Swagger UI
    e.GET("/swagger/*", echoSwagger.WrapHandler)
    
    // JSON документация
    e.GET("/swagger.json", func(c echo.Context) error {
        // Здесь можно отдавать сгенерированный swagger.json
        // или использовать встроенную генерацию
        return c.File("api/docs/swagger.json")
    })
    
    // YAML документация
    e.GET("/swagger.yaml", func(c echo.Context) error {
        return c.File("api/docs/swagger.yaml")
    })
}

// GenerateDocs генерирует документацию Swagger
func GenerateDocs() error {
    // Эта функция вызывается через Makefile
    // Реальная генерация выполняется через swag init
    return nil
}

// SwaggerInfo базовая информация для Swagger
func SwaggerInfo() map[string]interface{} {
    return map[string]interface{}{
        "title":       "Wishlist App API",
        "version":     "1.0.0",
        "description": "API для управления вишлистами",
        "termsOfService": "http://swagger.io/terms/",
        "contact": map[string]interface{}{
            "name":  "API Support",
            "url":   "http://www.swagger.io/support",
            "email": "support@swagger.io",
        },
        "license": map[string]interface{}{
            "name": "MIT",
            "url":  "https://opensource.org/licenses/MIT",
        },
    }
}
```

### 8. **Тесты с использованием testify и моками moq** (`test/integration/auth_test.go`)

```go
package integration_test

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "testing"
    "time"
    
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "wishlist-app/internal/app"
    "wishlist-app/internal/app/config"
    "wishlist-app/internal/domain/auth/delivery/http/dto"
    "wishlist-app/test/testutil"
)

// AuthIntegrationTestSuite набор тестов для аутентификации
type AuthIntegrationTestSuite struct {
    suite.Suite
    app        *app.App
    db         *pgxpool.Pool
    baseURL    string
    httpClient *http.Client
}

// SetupSuite настройка перед запуском набора тестов
func (s *AuthIntegrationTestSuite) SetupSuite() {
    // Загружаем тестовую конфигурацию
    cfg := &config.Config{
        App: config.AppConfig{
            Name:        "test",
            Environment: "test",
            Debug:       true,
        },
        Server: config.ServerConfig{
            Host: "localhost",
            Port: "0", // Используем случайный порт
        },
        Database: config.DatabaseConfig{
            Host:     "localhost",
            Port:     "5433", // Тестовый порт
            User:     "test",
            Password: "test",
            Name:     "wishlist_test",
            SSLMode:  "disable",
        },
        Auth: config.AuthConfig{
            JWTSecret:       "test-secret-key",
            AccessTokenTTL:  15 * time.Minute,
            RefreshTokenTTL: 7 * 24 * time.Hour,
        },
    }
    
    // Создаем приложение
    var err error
    s.app, err = app.New(cfg)
    assert.NoError(s.T(), err)
    
    // Запускаем приложение в горутине
    go func() {
        if err := s.app.Run(); err != nil {
            s.T().Logf("App run error: %v", err)
        }
    }()
    
    // Ждем запуска сервера
    time.Sleep(2 * time.Second)
    
    // Получаем URL сервера
    s.baseURL = "http://localhost:8080" // В реальности нужно получить из app
    
    s.httpClient = &http.Client{
        Timeout: 10 * time.Second,
    }
}

// TearDownSuite очистка после набора тестов
func (s *AuthIntegrationTestSuite) TearDownSuite() {
    if s.app != nil {
        s.app.Shutdown(context.Background())
    }
}

// SetupTest настройка перед каждым тестом
func (s *AuthIntegrationTestSuite) SetupTest() {
    // Очищаем базу данных перед каждым тестом
    err := testutil.CleanDatabase(s.db)
    assert.NoError(s.T(), err)
}

// TestRegister_Success тест успешной регистрации
func (s *AuthIntegrationTestSuite) TestRegister_Success() {
    req := dto.RegisterRequest{
        Email:    "test@example.com",
        Password: "Password123!",
        Name:     "Test User",
    }
    
    body, err := json.Marshal(req)
    assert.NoError(s.T(), err)
    
    resp, err := s.httpClient.Post(
        s.baseURL+"/api/v1/auth/register",
        "application/json",
        bytes.NewBuffer(body),
    )
    assert.NoError(s.T(), err)
    defer resp.Body.Close()
    
    assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)
    
    var response map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&response)
    assert.NoError(s.T(), err)
    
    assert.Equal(s.T(), "success", response["status"])
    assert.Contains(s.T(), response, "data")
    
    data := response["data"].(map[string]interface{})
    assert.Contains(s.T(), data, "access_token")
    assert.Contains(s.T(), data, "refresh_token")
    
    // Проверяем, что пользователь создан в БД
    user, err := testutil.GetUserByEmail(s.db, req.Email)
    assert.NoError(s.T(), err)
    assert.Equal(s.T(), req.Email, user.Email)
    assert.Equal(s.T(), req.Name, user.Name)
}

// TestRegister_DuplicateEmail тест регистрации с существующим email
func (s *AuthIntegrationTestSuite) TestRegister_DuplicateEmail() {
    // Сначала создаем пользователя
    err := testutil.CreateTestUser(s.db, "existing@example.com", "Password123!", "Existing User")
    assert.NoError(s.T(), err)
    
    req := dto.RegisterRequest{
        Email:    "existing@example.com",
        Password: "Password123!",
        Name:     "New User",
    }
    
    body, err := json.Marshal(req)
    assert.NoError(s.T(), err)
    
    resp, err := s.httpClient.Post(
        s.baseURL+"/api/v1/auth/register",
        "application/json",
        bytes.NewBuffer(body),
    )
    assert.NoError(s.T(), err)
    defer resp.Body.Close()
    
    assert.Equal(s.T(), http.StatusConflict, resp.StatusCode)
}

// TestLogin_Success тест успешного входа
func (s *AuthIntegrationTestSuite) TestLogin_Success() {
    // Создаем тестового пользователя
    err := testutil.CreateTestUser(s.db, "login@example.com", "Password123!", "Login User")
    assert.NoError(s.T(), err)
    
    req := dto.LoginRequest{
        Email:    "login@example.com",
        Password: "Password123!",
    }
    
    body, err := json.Marshal(req)
    assert.NoError(s.T(), err)
    
    resp, err := s.httpClient.Post(
        s.baseURL+"/api/v1/auth/login",
        "application/json",
        bytes.NewBuffer(body),
    )
    assert.NoError(s.T(), err)
    defer resp.Body.Close()
    
    assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
    
    var response map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&response)
    assert.NoError(s.T(), err)
    
    assert.Equal(s.T(), "success", response["status"])
}

// TestLogin_InvalidCredentials тест входа с неверными учетными данными
func (s *AuthIntegrationTestSuite) TestLogin_InvalidCredentials() {
    req := dto.LoginRequest{
        Email:    "nonexistent@example.com",
        Password: "WrongPassword123!",
    }
    
    body, err := json.Marshal(req)
    assert.NoError(s.T(), err)
    
    resp, err := s.httpClient.Post(
        s.baseURL+"/api/v1/auth/login",
        "application/json",
        bytes.NewBuffer(body),
    )
    assert.NoError(s.T(), err)
    defer resp.Body.Close()
    
    assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

// TestRefreshToken_Success тест успешного обновления токенов
func (s *AuthIntegrationTestSuite) TestRefreshToken_Success() {
    // Создаем пользователя и получаем refresh token
    user, err := testutil.CreateTestUser(s.db, "refresh@example.com", "Password123!", "Refresh User")
    assert.NoError(s.T(), err)
    
    // Логинимся, чтобы получить токены
    loginReq := dto.LoginRequest{
        Email:    "refresh@example.com",
        Password: "Password123!",
    }
    
    body, err := json.Marshal(loginReq)
    assert.NoError(s.T(), err)
    
    resp, err := s.httpClient.Post(
        s.baseURL+"/api/v1/auth/login",
        "application/json",
        bytes.NewBuffer(body),
    )
    assert.NoError(s.T(), err)
    defer resp.Body.Close()
    
    var loginResponse map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&loginResponse)
    assert.NoError(s.T(), err)
    
    loginData := loginResponse["data"].(map[string]interface{})
    refreshToken := loginData["refresh_token"].(string)
    
    // Обновляем токены
    refreshReq := dto.RefreshRequest{
        RefreshToken: refreshToken,
    }
    
    body, err = json.Marshal(refreshReq)
    assert.NoError(s.T(), err)
    
    resp, err = s.httpClient.Post(
        s.baseURL+"/api/v1/auth/refresh",
        "application/json",
        bytes.NewBuffer(body),
    )
    assert.NoError(s.T(), err)
    defer resp.Body.Close()
    
    assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
    
    var refreshResponse map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&refreshResponse)
    assert.NoError(s.T(), err)
    
    assert.Equal(s.T(), "success", refreshResponse["status"])
    
    // Проверяем, что создана новая сессия
    sessions, err := testutil.GetUserSessions(s.db, user.ID)
    assert.NoError(s.T(), err)
    assert.Len(s.T(), sessions, 1)
}

// TestGetProfile_Success тест получения профиля
func (s *AuthIntegrationTestSuite) TestGetProfile_Success() {
    // Создаем пользователя и получаем токен
    user, err := testutil.CreateTestUser(s.db, "profile@example.com", "Password123!", "Profile User")
    assert.NoError(s.T(), err)
    
    accessToken := testutil.GenerateTestToken(user.ID, user.Email)
    
    // Получаем профиль
    req, err := http.NewRequest("GET", s.baseURL+"/api/v1/auth/me", nil)
    assert.NoError(s.T(), err)
    
    req.Header.Set("Authorization", "Bearer "+accessToken)
    
    resp, err := s.httpClient.Do(req)
    assert.NoError(s.T(), err)
    defer resp.Body.Close()
    
    assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
    
    var response map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&response)
    assert.NoError(s.T(), err)
    
    assert.Equal(s.T(), "success", response["status"])
    
    data := response["data"].(map[string]interface{})
    assert.Equal(s.T(), user.ID, data["id"])
    assert.Equal(s.T(), user.Email, data["email"])
    assert.Equal(s.T(), user.Name, data["name"])
}

// TestGetProfile_Unauthorized тест получения профиля без авторизации
func (s *AuthIntegrationTestSuite) TestGetProfile_Unauthorized() {
    req, err := http.NewRequest("GET", s.baseURL+"/api/v1/auth/me", nil)
    assert.NoError(s.T(), err)
    
    resp, err := s.httpClient.Do(req)
    assert.NoError(s.T(), err)
    defer resp.Body.Close()
    
    assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

// RunTests запускает все тесты
func TestAuthIntegrationTestSuite(t *testing.T) {
    if testing.Short() {
        t.Skip("Пропускаем интеграционные тесты в режиме short")
    }
    
    suite.Run(t, new(AuthIntegrationTestSuite))
}
```

### 9. **Моки с использованием moq** (`test/mocks/auth_repository_mock.go`)

```go
// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
    "context"
    "sync"
    "wishlist-app/internal/domain/auth/models"
    "wishlist-app/internal/domain/auth/repository"
)

// Ensure, that AuthRepositoryMock does implement repository.AuthRepository.
var _ repository.AuthRepository = &AuthRepositoryMock{}

// AuthRepositoryMock is a mock implementation of repository.AuthRepository.
type AuthRepositoryMock struct {
    // CreateUserFunc mocks the CreateUser method.
    CreateUserFunc func(ctx context.Context, user *models.User) error

    // GetUserByIDFunc mocks the GetUserByID method.
    GetUserByIDFunc func(ctx context.Context, id string) (*models.User, error)

    // GetUserByEmailFunc mocks the GetUserByEmail method.
    GetUserByEmailFunc func(ctx context.Context, email string) (*models.User, error)

    // UpdateUserFunc mocks the UpdateUser method.
    UpdateUserFunc func(ctx context.Context, user *models.User) error

    // DeleteUserFunc mocks the DeleteUser method.
    DeleteUserFunc func(ctx context.Context, id string) error

    // CreateSessionFunc mocks the CreateSession method.
    CreateSessionFunc func(ctx context.Context, session *models.Session) error

    // GetSessionByRefreshTokenFunc mocks the GetSessionByRefreshToken method.
    GetSessionByRefreshTokenFunc func(ctx context.Context, refreshToken string) (*models.Session, error)

    // DeleteSessionFunc mocks the DeleteSession method.
    DeleteSessionFunc func(ctx context.Context, id string) error

    // calls tracks calls to the methods.
    calls struct {
        // CreateUser holds details about calls to the CreateUser method.
        CreateUser []struct {
            // Ctx is the ctx argument value.
            Ctx context.Context
            // User is the user argument value.
            User *models.User
        }
        // GetUserByID holds details about calls to the GetUserByID method.
        GetUserByID []struct {
            // Ctx is the ctx argument value.
            Ctx context.Context
            // ID is the id argument value.
            ID string
        }
        // GetUserByEmail holds details about calls to the GetUserByEmail method.
        GetUserByEmail []struct {
            // Ctx is the ctx argument value.
            Ctx context.Context
            // Email is the email argument value.
            Email string
        }
        // UpdateUser holds details about calls to the UpdateUser method.
        UpdateUser []struct {
            // Ctx is the ctx argument value.
            Ctx context.Context
            // User is the user argument value.
            User *models.User
        }
        // DeleteUser holds details about calls to the DeleteUser method.
        DeleteUser []struct {
            // Ctx is the ctx argument value.
            Ctx context.Context
            // ID is the id argument value.
            ID string
        }
        // CreateSession holds details about calls to the CreateSession method.
        CreateSession []struct {
            // Ctx is the ctx argument value.
            Ctx context.Context
            // Session is the session argument value.
            Session *models.Session
        }
        // GetSessionByRefreshToken holds details about calls to the GetSessionByRefreshToken method.
        GetSessionByRefreshToken []struct {
            // Ctx is the ctx argument value.
            Ctx context.Context
            // RefreshToken is the refreshToken argument value.
            RefreshToken string
        }
        // DeleteSession holds details about calls to the DeleteSession method.
        DeleteSession []struct {
            // Ctx is the ctx argument value.
            Ctx context.Context
            // ID is the id argument value.
            ID string
        }
    }
    lockCreateUser                sync.RWMutex
    lockGetUserByID               sync.RWMutex
    lockGetUserByEmail            sync.RWMutex
    lockUpdateUser                sync.RWMutex
    lockDeleteUser                sync.RWMutex
    lockCreateSession             sync.RWMutex
    lockGetSessionByRefreshToken  sync.RWMutex
    lockDeleteSession             sync.RWMutex
}

// CreateUser calls CreateUserFunc.
func (mock *AuthRepositoryMock) CreateUser(ctx context.Context, user *models.User) error {
    if mock.CreateUserFunc == nil {
        panic("AuthRepositoryMock.CreateUserFunc: method is nil but AuthRepository.CreateUser was just called")
    }
    callInfo := struct {
        Ctx  context.Context
        User *models.User
    }{
        Ctx:  ctx,
        User: user,
    }
    mock.lockCreateUser.Lock()
    mock.calls.CreateUser = append(mock.calls.CreateUser, callInfo)
    mock.lockCreateUser.Unlock()
    return mock.CreateUserFunc(ctx, user)
}

// CreateUserCalls gets all the calls that were made to CreateUser.
// Check the length with:
//     len(mockedAuthRepository.CreateUserCalls())
func (mock *AuthRepositoryMock) CreateUserCalls() []struct {
    Ctx  context.Context
    User *models.User
} {
    var calls []struct {
        Ctx  context.Context
        User *models.User
    }
    mock.lockCreateUser.RLock()
    calls = mock.calls.CreateUser
    mock.lockCreateUser.RUnlock()
    return calls
}

// GetUserByID calls GetUserByIDFunc.
func (mock *AuthRepositoryMock) GetUserByID(ctx context.Context, id string) (*models.User, error) {
    if mock.GetUserByIDFunc == nil {
        panic("AuthRepositoryMock.GetUserByIDFunc: method is nil but AuthRepository.GetUserByID was just called")
    }
    callInfo := struct {
        Ctx context.Context
        ID  string
    }{
        Ctx: ctx,
        ID:  id,
    }
    mock.lockGetUserByID.Lock()
    mock.calls.GetUserByID = append(mock.calls.GetUserByID, callInfo)
    mock.lockGetUserByID.Unlock()
    return mock.GetUserByIDFunc(ctx, id)
}

// GetUserByIDCalls gets all the calls that were made to GetUserByID.
// Check the length with:
//     len(mockedAuthRepository.GetUserByIDCalls())
func (mock *AuthRepositoryMock) GetUserByIDCalls() []struct {
    Ctx context.Context
    ID  string
} {
    var calls []struct {
        Ctx context.Context
        ID  string
    }
    mock.lockGetUserByID.RLock()
    calls = mock.calls.GetUserByID
    mock.lockGetUserByID.RUnlock()
    return calls
}

// GetUserByEmail calls GetUserByEmailFunc.
func (mock *AuthRepositoryMock) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
    if mock.GetUserByEmailFunc == nil {
        panic("AuthRepositoryMock.GetUserByEmailFunc: method is nil but AuthRepository.GetUserByEmail was just called")
    }
    callInfo := struct {
        Ctx   context.Context
        Email string
    }{
        Ctx:   ctx,
        Email: email,
    }
    mock.lockGetUserByEmail.Lock()
    mock.calls.GetUserByEmail = append(mock.calls.GetUserByEmail, callInfo)
    mock.lockGetUserByEmail.Unlock()
    return mock.GetUserByEmailFunc(ctx, email)
}

// GetUserByEmailCalls gets all the calls that were made to GetUserByEmail.
// Check the length with:
//     len(mockedAuthRepository.GetUserByEmailCalls())
func (mock *AuthRepositoryMock) GetUserByEmailCalls() []struct {
    Ctx   context.Context
    Email string
} {
    var calls []struct {
        Ctx   context.Context
        Email string
    }
    mock.lockGetUserByEmail.RLock()
    calls = mock.calls.GetUserByEmail
    mock.lockGetUserByEmail.RUnlock()
    return calls
}

// UpdateUser calls UpdateUserFunc.
func (mock *AuthRepositoryMock) UpdateUser(ctx context.Context, user *models.User) error {
    if mock.UpdateUserFunc == nil {
        panic("AuthRepositoryMock.UpdateUserFunc: method is nil but AuthRepository.UpdateUser was just called")
    }
    callInfo := struct {
        Ctx  context.Context
        User *models.User
    }{
        Ctx:  ctx,
        User: user,
    }
    mock.lockUpdateUser.Lock()
    mock.calls.UpdateUser = append(mock.calls.UpdateUser, callInfo)
    mock.lockUpdateUser.Unlock()
    return mock.UpdateUserFunc(ctx, user)
}

// UpdateUserCalls gets all the calls that were made to UpdateUser.
// Check the length with:
//     len(mockedAuthRepository.UpdateUserCalls())
func (mock *AuthRepositoryMock) UpdateUserCalls() []struct {
    Ctx  context.Context
    User *models.User
} {
    var calls []struct {
        Ctx  context.Context
        User *models.User
    }
    mock.lockUpdateUser.RLock()
    calls = mock.calls.UpdateUser
    mock.lockUpdateUser.RUnlock()
    return calls
}

// DeleteUser calls DeleteUserFunc.
func (mock *AuthRepositoryMock) DeleteUser(ctx context.Context, id string) error {
    if mock.DeleteUserFunc == nil {
        panic("AuthRepositoryMock.DeleteUserFunc: method is nil but AuthRepository.DeleteUser was just called")
    }
    callInfo := struct {
        Ctx context.Context
        ID  string
    }{
        Ctx: ctx,
        ID:  id,
    }
    mock.lockDeleteUser.Lock()
    mock.calls.DeleteUser = append(mock.calls.DeleteUser, callInfo)
    mock.lockDeleteUser.Unlock()
    return mock.DeleteUserFunc(ctx, id)
}

// DeleteUserCalls gets all the calls that were made to DeleteUser.
// Check the length with:
//     len(mockedAuthRepository.DeleteUserCalls())
func (mock *AuthRepositoryMock) DeleteUserCalls() []struct {
    Ctx context.Context
    ID  string
} {
    var calls []struct {
        Ctx context.Context
        ID  string
    }
    mock.lockDeleteUser.RLock()
    calls = mock.calls.DeleteUser
    mock.lockDeleteUser.RUnlock()
    return calls
}

// CreateSession calls CreateSessionFunc.
func (mock *AuthRepositoryMock) CreateSession(ctx context.Context, session *models.Session) error {
    if mock.CreateSessionFunc == nil {
        panic("AuthRepositoryMock.CreateSessionFunc: method is nil but AuthRepository.CreateSession was just called")
    }
    callInfo := struct {
        Ctx     context.Context
        Session *models.Session
    }{
        Ctx:     ctx,
        Session: session,
    }
    mock.lockCreateSession.Lock()
    mock.calls.CreateSession = append(mock.calls.CreateSession, callInfo)
    mock.lockCreateSession.Unlock()
    return mock.CreateSessionFunc(ctx, session)
}

// CreateSessionCalls gets all the calls that were made to CreateSession.
// Check the length with:
//     len(mockedAuthRepository.CreateSessionCalls())
func (mock *AuthRepositoryMock) CreateSessionCalls() []struct {
    Ctx     context.Context
    Session *models.Session
} {
    var calls []struct {
        Ctx     context.Context
        Session *models.Session
    }
    mock.lockCreateSession.RLock()
    calls = mock.calls.CreateSession
    mock.lockCreateSession.RUnlock()
    return calls
}

// GetSessionByRefreshToken calls GetSessionByRefreshTokenFunc.
func (mock *AuthRepositoryMock) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
    if mock.GetSessionByRefreshTokenFunc == nil {
        panic("AuthRepositoryMock.GetSessionByRefreshTokenFunc: method is nil but AuthRepository.GetSessionByRefreshToken was just called")
    }
    callInfo := struct {
        Ctx          context.Context
        RefreshToken string
    }{
        Ctx:          ctx,
        RefreshToken: refreshToken,
    }
    mock.lockGetSessionByRefreshToken.Lock()
    mock.calls.GetSessionByRefreshToken = append(mock.calls.GetSessionByRefreshToken, callInfo)
    mock.lockGetSessionByRefreshToken.Unlock()
    return mock.GetSessionByRefreshTokenFunc(ctx, refreshToken)
}

// GetSessionByRefreshTokenCalls gets all the calls that were made to GetSessionByRefreshToken.
// Check the length with:
//     len(mockedAuthRepository.GetSessionByRefreshTokenCalls())
func (mock *AuthRepositoryMock) GetSessionByRefreshTokenCalls() []struct {
    Ctx          context.Context
    RefreshToken string
} {
    var calls []struct {
        Ctx          context.Context
        RefreshToken string
    }
    mock.lockGetSessionByRefreshToken.RLock()
    calls = mock.calls.GetSessionByRefreshToken
    mock.lockGetSessionByRefreshToken.RUnlock()
    return calls
}

// DeleteSession calls DeleteSessionFunc.
func (mock *AuthRepositoryMock) DeleteSession(ctx context.Context, id string) error {
    if mock.DeleteSessionFunc == nil {
        panic("AuthRepositoryMock.DeleteSessionFunc: method is nil but AuthRepository.DeleteSession was just called")
    }
    callInfo := struct {
        Ctx context.Context
        ID  string
    }{
        Ctx: ctx,
        ID:  id,
    }
    mock.lockDeleteSession.Lock()
    mock.calls.DeleteSession = append(mock.calls.DeleteSession, callInfo)
    mock.lockDeleteSession.Unlock()
    return mock.DeleteSessionFunc(ctx, id)
}

// DeleteSessionCalls gets all the calls that were made to DeleteSession.
// Check the length with:
//     len(mockedAuthRepository.DeleteSessionCalls())
func (mock *AuthRepositoryMock) DeleteSessionCalls() []struct {
    Ctx context.Context
    ID  string
} {
    var calls []struct {
        Ctx context.Context
        ID  string
    }
    mock.lockDeleteSession.RLock()
    calls = mock.calls.DeleteSession
    mock.lockDeleteSession.RUnlock()
    return calls
}
```

### 10. **Docker Compose для деплоя** (`deployments/docker-compose.yml`)

```yaml
version: '3.8'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:15-alpine
    container_name: wishlist-postgres
    environment:
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-postgres}
      POSTGRES_DB: ${DB_NAME:-wishlist_db}
      POSTGRES_INITDB_ARGS: "--encoding=UTF8 --lc-collate=C --lc-ctype=C"
    ports:
      - "${DB_PORT:-5432}:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    networks:
      - wishlist-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-postgres}"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  # Application
  app:
    build:
      context: ..
      dockerfile: deployments/docker/Dockerfile
      args:
        APP_VERSION: ${APP_VERSION:-1.0.0}
    container_name: wishlist-app
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      # Application
      APP_NAME: ${APP_NAME:-wishlist-app}
      APP_VERSION: ${APP_VERSION:-1.0.0}
      APP_ENV: ${APP_ENV:-production}
      APP_DEBUG: ${APP_DEBUG:-false}
      
      # Server
      HOST: ${HOST:-0.0.0.0}
      PORT: ${PORT:-8080}
      SERVER_READ_TIMEOUT: ${SERVER_READ_TIMEOUT:-30s}
      SERVER_WRITE_TIMEOUT: ${SERVER_WRITE_TIMEOUT:-30s}
      SERVER_IDLE_TIMEOUT: ${SERVER_IDLE_TIMEOUT:-60s}
      
      # Database
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: ${DB_USER:-postgres}
      DB_PASSWORD: ${DB_PASSWORD:-postgres}
      DB_NAME: ${DB_NAME:-wishlist_db}
      DB_SSL_MODE: ${DB_SSL_MODE:-disable}
      DB_MAX_OPEN_CONNS: ${DB_MAX_OPEN_CONNS:-25}
      DB_MAX_IDLE_CONNS: ${DB_MAX_IDLE_CONNS:-25}
      DB_CONN_MAX_LIFETIME: ${DB_CONN_MAX_LIFETIME:-5m}
      DB_MIGRATIONS_PATH: /app/migrations
      DB_AUTO_MIGRATE: ${DB_AUTO_MIGRATE:-true}
      
      # Authentication
      JWT_SECRET: ${JWT_SECRET}
      JWT_ACCESS_TOKEN_TTL: ${JWT_ACCESS_TOKEN_TTL:-15m}
      JWT_REFRESH_TOKEN_TTL: ${JWT_REFRESH_TOKEN_TTL:-7d}
      JWT_ISSUER: ${JWT_ISSUER:-wishlist-app}
      JWT_AUDIENCE: ${JWT_AUDIENCE:-wishlist-app}
      
      # Logging
      LOG_LEVEL: ${LOG_LEVEL:-info}
      LOG_FORMAT: ${LOG_FORMAT:-json}
      LOG_OUTPUT: ${LOG_OUTPUT:-stdout}
      
      # Swagger
      SWAGGER_ENABLED: ${SWAGGER_ENABLED:-false}
      SWAGGER_HOST: ${SWAGGER_HOST:-localhost:8080}
    ports:
      - "${PORT:-8080}:8080"
    volumes:
      - ./storage:/app/storage
      - ./migrations:/app/migrations
    networks:
      - wishlist-network
    restart: unless-stopped
    command: >
      sh -c "
        echo 'Waiting for database to be ready...' &&
        sleep 5 &&
        echo 'Running migrations...' &&
        migrate -path /app/migrations -database 'postgres://${DB_USER:-postgres}:${DB_PASSWORD:-postgres}@postgres:5432/${DB_NAME:-wishlist_db}?sslmode=${DB_SSL_MODE:-disable}' up &&
        echo 'Starting application...' &&
        /app/wishlist-app
      "

  # Nginx (Reverse Proxy)
  nginx:
    image: nginx:alpine
    container_name: wishlist-nginx
    depends_on:
      - app
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
      - ./logs/nginx:/var/log/nginx
    networks:
      - wishlist-network
    restart: unless-stopped

networks:
  wishlist-network:
    driver: bridge

volumes:
  postgres_data:
    driver: local
```

### 11. **Dockerfile для продакшена** (`deployments/docker/Dockerfile`)

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git gcc musl-dev

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Устанавливаем инструменты для миграций
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Генерируем Swagger документацию
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/api/main.go -o api/docs

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=${APP_VERSION:-1.0.0}" \
    -o wishlist-app cmd/api/main.go

# Production stage
FROM alpine:latest

# Устанавливаем зависимости для runtime
RUN apk add --no-cache ca-certificates tzdata

# Создаем не-root пользователя
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

# Создаем рабочие директории
RUN mkdir -p /app /app/storage /app/migrations && \
    chown -R appuser:appuser /app

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем бинарник из builder stage
COPY --from=builder /app/wishlist-app .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/api/docs ./api/docs

# Копируем конфигурационные файлы
COPY deployments/docker/configs/ ./configs/

# Устанавливаем права
RUN chown -R appuser:appuser /app && \
    chmod +x /app/wishlist-app

# Переключаемся на не-root пользователя
USER appuser

# Экспортируем порт
EXPOSE 8080

# Команда запуска
CMD ["/app/wishlist-app"]
```

### 12. **Makefile с обновленными командами**

```makefile
.PHONY: help run build test migrate-up migrate-down migrate-create \
        docker-up docker-down docker-build swag mocks lint

# Colors
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m

# Variables
APP_NAME=wishlist-app
APP_VERSION=1.0.0
GO_VERSION=1.21

help:
	@echo ""
	@echo "${YELLOW}${APP_NAME} - Makefile Commands${NC}"
	@echo ""
	@echo "${GREEN}Development:${NC}"
	@echo "  ${BLUE}make run${NC}         - Run the application locally"
	@echo "  ${BLUE}make watch${NC}       - Run with hot reload (air)"
	@echo "  ${BLUE}make build${NC}       - Build the application"
	@echo ""
	@echo "${GREEN}Testing:${NC}"
	@echo "  ${BLUE}make test${NC}        - Run all tests"
	@echo "  ${BLUE}make test-unit${NC}   - Run unit tests"
	@echo "  ${BLUE}make test-integration${NC} - Run integration tests"
	@echo ""
	@echo "${GREEN}Database:${NC}"
	@echo "  ${BLUE}make migrate-up${NC}  - Run database migrations"
	@echo "  ${BLUE}make migrate-down${NC} - Rollback database migrations"
	@echo "  ${BLUE}make migrate-create${NC} - Create new migration"
	@echo ""
	@echo "${GREEN}Docker:${NC}"
	@echo "  ${BLUE}make docker-up${NC}   - Start Docker containers"
	@echo "  ${BLUE}make docker-down${NC} - Stop Docker containers"
	@echo "  ${BLUE}make docker-build${NC} - Build Docker image"
	@echo ""
	@echo "${GREEN}Code Generation:${NC}"
	@echo "  ${BLUE}make swag${NC}        - Generate Swagger documentation"
	@echo "  ${BLUE}make mocks${NC}       - Generate mocks for tests"
	@echo ""
	@echo "${GREEN}Code Quality:${NC}"
	@echo "  ${BLUE}make lint${NC}        - Run linter"
	@echo "  ${BLUE}make format${NC}      - Format code"
	@echo "  ${BLUE}make tidy${NC}        - Tidy go modules"
	@echo ""

run:
	@echo "${YELLOW}Starting application...${NC}"
	@go run cmd/api/main.go

watch:
	@echo "${YELLOW}Starting application with hot reload...${NC}"
	@air -c .air.toml

build:
	@echo "${YELLOW}Building application...${NC}"
	@CGO_ENABLED=0 go build \
		-ldflags="-w -s -X main.version=${APP_VERSION}" \
		-o bin/${APP_NAME} cmd/api/main.go
	@echo "${GREEN}Build completed: bin/${APP_NAME}${NC}"

test:
	@echo "${YELLOW}Running all tests...${NC}"
	@go test ./... -v

test-unit:
	@echo "${YELLOW}Running unit tests...${NC}"
	@go test ./internal/domain/... -v -short

test-integration:
	@echo "${YELLOW}Running integration tests...${NC}"
	@go test ./test/integration/... -v

migrate-up:
	@echo "${YELLOW}Running database migrations...${NC}"
	@migrate -path ./migrations \
		-database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}" \
		up

migrate-down:
	@echo "${YELLOW}Rolling back database migrations...${NC}"
	@migrate -path ./migrations \
		-database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}" \
		down 1

migrate-create:
	@echo "${YELLOW}Creating new migration...${NC}"
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir ./migrations -seq $${name// /_}

docker-up:
	@echo "${YELLOW}Starting Docker containers...${NC}"
	@docker-compose -f deployments/docker-compose.yml up -d
	@echo "${GREEN}Containers started!${NC}"

docker-down:
	@echo "${YELLOW}Stopping Docker containers...${NC}"
	@docker-compose -f deployments/docker-compose.yml down
	@echo "${GREEN}Containers stopped!${NC}"

docker-build:
	@echo "${YELLOW}Building Docker image...${NC}"
	@docker build \
		-t ${APP_NAME}:${APP_VERSION} \
		-t ${APP_NAME}:latest \
		-f deployments/docker/Dockerfile .
	@echo "${GREEN}Docker image built: ${APP_NAME}:${APP_VERSION}${NC}"

swag:
	@echo "${YELLOW}Generating Swagger documentation...${NC}"
	@swag init -g cmd/api/main.go -o api/docs
	@echo "${GREEN}Swagger documentation generated!${NC}"

mocks:
	@echo "${YELLOW}Generating mocks...${NC}"
	@go generate ./...
	@echo "${GREEN}Mocks generated!${NC}"

lint:
	@echo "${YELLOW}Running linter...${NC}"
	@golangci-lint run ./...

format:
	@echo "${YELLOW}Formatting code...${NC}"
	@gofmt -w -s .
	@goimports -w .

tidy:
	@echo "${YELLOW}Tidying go modules...${NC}"
	@go mod tidy
	@go mod verify

clean:
	@echo "${YELLOW}Cleaning build artifacts...${NC}"
	@rm -rf bin/ coverage.out coverage.html
	@go clean
	@echo "${GREEN}Clean completed!${NC}"
```

### 13. **go.mod с обновленными зависимостями**

```mod
module wishlist-app

go 1.21

require (
    // Web Framework
    github.com/labstack/echo/v4 v4.11.1
    
    // Environment Variables
    github.com/joho/godotenv v1.5.1
    
    // Swagger Documentation
    github.com/swaggo/echo-swagger v1.4.1
    github.com/swaggo/swag v1.16.2
    
    // Validation
    github.com/go-playground/validator/v10 v10.15.5
    
    // Database
    github.com/jackc/pgx/v5 v5.5.0
    github.com/jmoiron/sqlx v1.3.5
    
    // Migrations
    github.com/golang-migrate/migrate/v4 v4.16.2
    
    // Testing
    github.com/stretchr/testify v1.8.4
    github.com/matryer/moq v0.3.2
    
    // JWT
    github.com/golang-jwt/jwt/v5 v5.0.0
    
    // Logging
    go.uber.org/zap v1.26.0
    
    // Configuration
    github.com/spf13/viper v1.17.0
    
    // Hot Reload (development only)
    github.com/cosmtrek/air v1.49.0
)

require (
    // Transitive dependencies
    github.com/KyleBanks/depth v1.2.1 // indirect
    github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
    github.com/fsnotify/fsnotify v1.6.0 // indirect
    github.com/ghodss/yaml v1.0.0 // indirect
    github.com/go-openapi/jsonpointer v0.20.0 // indirect
    github.com/go-openapi/jsonreference v0.20.2 // indirect
    github.com/go-openapi/spec v0.20.9 // indirect
    github.com/go-openapi/swag v0.22.4 // indirect
    github.com/go-playground/locales v0.14.1 // indirect
    github.com/go-playground/universal-translator v0.18.1 // indirect
    github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
    github.com/hashicorp/errwrap v1.1.0 // indirect
    github.com/hashicorp/go-multierror v1.1.1 // indirect
    github.com/hashicorp/hcl v1.0.0 // indirect
    github.com/imdario/mergo v0.3.16 // indirect
    github.com/jackc/pgpassfile v1.0.0 // indirect
    github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
    github.com/jackc/puddle/v2 v2.2.1 // indirect
    github.com/josharian/intern v1.0.0 // indirect
    github.com/labstack/gommon v0.4.0 // indirect
    github.com/leodido/go-urn v1.2.4 // indirect
    github.com/magiconair/properties v1.8.7 // indirect
    github.com/mailru/easyjson v0.7.7 // indirect
    github.com/mattn/go-colorable v0.1.13 // indirect
    github.com/mattn/go-isatty v0.0.19 // indirect
    github.com/mitchellh/mapstructure v1.5.0 // indirect
    github.com/pelletier/go-toml/v2 v2.1.0 // indirect
    github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
    github.com/rogpeppe/go-internal v1.11.0 // indirect
    github.com/sagikazarmark/locafero v0.3.0 // indirect
    github.com/sagikazarmark/slog-shim v0.1.0 // indirect
    github.com/sourcegraph/conc v0.3.0 // indirect
    github.com/spf13/afero v1.10.0 // indirect
    github.com/spf13/cast v1.5.1 // indirect
    github.com/spf13/pflag v1.0.5 // indirect
    github.com/subosito/gotenv v1.6.0 // indirect
    github.com/valyala/bytebufferpool v1.0.0 // indirect
    github.com/valyala/fasttemplate v1.2.2 // indirect
    go.uber.org/atomic v1.9.0 // indirect
    go.uber.org/multierr v1.11.0 // indirect
    golang.org/x/crypto v0.14.0 // indirect
    golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
    golang.org/x/net v0.17.0 // indirect
    golang.org/x/sync v0.3.0 // indirect
    golang.org/x/sys v0.13.0 // indirect
    golang.org/x/text v0.13.0 // indirect
    golang.org/x/time v0.3.0 // indirect
    golang.org/x/tools v0.14.0 // indirect
    gopkg.in/ini.v1 v1.67.0 // indirect
    gopkg.in/yaml.v2 v2.4.0 // indirect
    gopkg.in/yaml.v3 v3.0.1 // indirect
)
```
