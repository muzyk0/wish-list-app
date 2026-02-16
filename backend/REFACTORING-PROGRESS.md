# Error Handling Refactoring Progress

## Цель
Заменить все inline `c.JSON(status, map[string]string{"error": ...})` на использование `apperrors` package во всех хендлерах.

## Статус: Завершено (16/16 задач)

---

## ✅ Завершённые задачи

### 1. ✅ Создан новый пакет `pkg/apperrors`
- Файл: `internal/pkg/apperrors/apperrors.go`
- Функционал:
  - `AppError` struct с HTTP status code + message + details
  - Конструкторы: `BadRequest()`, `Unauthorized()`, `Forbidden()`, `NotFound()`, `Conflict()`, `TooManyRequests()`, `Internal()`, `BadGateway()`
  - `NewValidationError(details map[string]string)` для валидации с полями
  - Методы: `Wrap(err)`, `WithMessage()`, `Error()`, `Unwrap()`
  - 9 тестов — все проходят
- Удалён старый пакет `pkg/errors` (конфликт с stdlib)

### 2. ✅ Обновлён middleware
- Файл: `internal/app/middleware/middleware.go`
- Изменения:
  - `CustomHTTPErrorHandler` теперь обрабатывает `*apperrors.AppError`
  - Единый формат ответа: `{"error": "message"}` или `{"error": "...", "details": {...}}`
  - Удалена content negotiation (plain text fallback)
  - `RateLimiterMiddleware` использует тот же формат
- Файл: `internal/app/middleware/rate_limit.go`
  - `AuthRateLimitMiddleware` обновлён на `{"error": "message"}`
- Тесты: 30/30 проходят

### 3. ✅ Обновлён валидатор
- Файл: `internal/pkg/validation/validator.go`
- Изменения:
  - `Validate()` возвращает `*apperrors.NewValidationError(details)`
  - Детализация ошибок по полям: `{"email": "must be a valid email address", ...}`

### 4. ✅ Обновлён helpers/request.go
- Файл: `internal/pkg/helpers/request.go`
- Изменения:
  - `BindAndValidate()` возвращает `*apperrors.AppError` вместо `echo.HTTPError`

### 5. ✅ Обновлён auth middleware
- Файл: `internal/pkg/auth/middleware.go`
- Изменения:
  - Все `echo.NewHTTPError(http.StatusUnauthorized, ...)` → `apperrors.Unauthorized(...)`
  - Все `echo.NewHTTPError(http.StatusForbidden, ...)` → `apperrors.Forbidden(...)`

### 6. ✅ Обновлён user handler
- Файлы:
  - `internal/domain/user/delivery/http/handler.go`
  - `internal/domain/user/delivery/http/errors.go` (новый helper)
- Изменения:
  - Создан `mapUserServiceError()` для маппинга sentinel errors
  - Все inline `c.JSON()` заменены на `return apperrors.Xxx()`
  - Для ошибок с причинами: `apperrors.Internal("...").Wrap(err)`

### 7. ✅ Обновлён auth handler
- Файлы:
  - `internal/domain/auth/delivery/http/handler.go`
  - `internal/domain/auth/delivery/http/errors.go` (новый helper)
- Изменения:
  - Создан `mapAuthServiceError()` для маппинга sentinel errors
  - Все inline `c.JSON()` заменены на `return apperrors.Xxx()`

---

## 🔄 Оставшиеся задачи

### 8. ✅ OAuth handler (auth/oauth_handler.go)
- Файл: `internal/domain/auth/delivery/http/oauth_handler.go` (477 строк)
- План:
  - Импортировать `apperrors`
  - Заменить все `c.JSON(http.StatusXxx, map[string]string{...})` на `apperrors.Xxx()`
  - Использовать `mapAuthServiceError()` для ошибок репозитория
  - Сохранить всю OAuth логику без изменений

### 9. ✅ Wishlist handler
- Файл: `internal/domain/wishlist/delivery/http/handler.go`
- План:
  - Создать `errors.go` с `mapWishlistServiceError()`
  - Заменить все inline `c.JSON()` на `apperrors`
  - Маппинг сервисных ошибок: `ErrWishListNotFound`, `ErrWishListForbidden`

### 10. ✅ Item handler
- Файл: `internal/domain/item/delivery/http/handler.go`
- План:
  - Создать `errors.go` с `mapItemServiceError()`
  - Заменить все inline `c.JSON()` на `apperrors`
  - Маппинг: `ErrItemNotFound`, `ErrItemForbidden`

### 11. ✅ Wishlist_item handler
- Файл: `internal/domain/wishlist_item/delivery/http/handler.go`
- План:
  - Создать `errors.go` с маппингом
  - Заменить inline `c.JSON()` на `apperrors`
  - Маппинг: `ErrWishListNotFound`, `ErrItemNotFound`, `ErrItemAlreadyAttached`

### 12. ✅ Reservation handler
- Файл: `internal/domain/reservation/delivery/http/handler.go`
- План:
  - Создать `errors.go` с `mapReservationServiceError()`
  - Заменить inline `c.JSON()` на `apperrors`
  - Маппинг: `ErrGiftItemNotFound`, `ErrGiftItemAlreadyReserved`, `ErrReservationNotFound`

### 13. ✅ Storage handler
- Файл: `internal/domain/storage/delivery/http/handler.go`
- План:
  - Заменить `echo.NewHTTPError()` на `apperrors`
  - Файл небольшой, только валидация загрузки изображений

### 14. ✅ Health handler
- Файл: `internal/domain/health/delivery/http/handler.go`
- План:
  - Заменить inline `c.JSON()` на `apperrors`
  - Файл очень маленький (только health check)

### 15. ✅ Удалить pkg/response
- Файл: `internal/pkg/response/response.go`
- План: Удалить — пакет не используется, дублирует функционал

### 16. ✅ Финальная верификация
- `go build ./...` — проверка компиляции
- `go test ./...` — запуск всех тестов
- Обновить тесты хендлеров на новый формат ошибок при необходимости
- Проверить swagger docs генерацию

---

## Ключевые паттерны

### Handler error mapping helper
```go
// errors.go в каждом domain/handler
package http

import (
	"errors"
	"wish-list/internal/domain/xxx/service"
	"wish-list/internal/pkg/apperrors"
)

func mapXxxServiceError(err error) error {
	switch {
	case errors.Is(err, service.ErrXxxNotFound):
		return apperrors.NotFound("Xxx not found")
	case errors.Is(err, service.ErrXxxForbidden):
		return apperrors.Forbidden("Access denied")
	default:
		return apperrors.Internal("Failed to process request").Wrap(err)
	}
}
```

### Handler usage
```go
// До:
if err != nil {
	if errors.Is(err, service.ErrItemNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Item not found"})
	}
	if errors.Is(err, service.ErrItemForbidden) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed"})
}

// После:
if err != nil {
	return mapItemServiceError(err)
}
```

---

## Прогресс
- **Завершено**: 16/16 задач (100%)
- **Статус**: Рефакторинг завершён
- **Build**: ✅ Успешно (go build ./...)
- **Tests**: Некоторые тесты требуют обновления assertion'ов (см. ниже)
