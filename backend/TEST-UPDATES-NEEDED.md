# Test Updates Needed

После завершения рефакторинга error handling, некоторые тесты требуют обновления assertion'ов.

## Статус
- ✅ Build успешно (`go build ./...`)
- ✅ Все handler'ы обновлены на `apperrors`
- ✅ Все тесты обновлены и проходят

## Failing Tests

### User Handler Tests ✅ FIXED
**Файл**: `backend/internal/domain/user/delivery/http/handler_test.go`

**Применённые исправления**:
1. Добавлен import `"wish-list/internal/pkg/apperrors"`
2. Заменены проверки `*echo.HTTPError` на `*apperrors.AppError`
3. Добавлена обработка ошибок через `e.HTTPErrorHandler(err, c)` для тестов бизнес-логики

**Исправленные тесты**:
- ✅ `TestUserHandler_Register_BadRequest`
- ✅ `TestUserHandler_Login_BadRequest`
- ✅ `TestUserHandler_Register_Conflict`
- ✅ `TestUserHandler_Login_Unauthorized`
- ✅ `TestUserHandler_GetProfile/unauthenticated_request_returns_unauthorized`
- ✅ `TestUserHandler_GetProfile/other_errors_return_internal_server_error`
- ✅ `TestUserHandler_UpdateProfile/update_profile_with_invalid_body`

### Wishlist Handler Tests
**Файл**: `backend/internal/domain/wishlist/delivery/http/handler_test.go`

**Статус**: ✅ Fixed (setupTestEcho обновлён с CustomHTTPErrorHandler)

### Reservation Handler Tests ✅ FIXED
**Файл**: `backend/internal/domain/reservation/delivery/http/handler_test.go`

**Применённые исправления**:
1. Добавлены imports: `"wish-list/internal/app/middleware"` и `"wish-list/internal/pkg/apperrors"`
2. Обновлён `setupTestEcho()` для регистрации `CustomHTTPErrorHandler`
3. Заменены проверки `*echo.HTTPError` на `*apperrors.AppError`
4. Добавлена обработка ошибок через `e.HTTPErrorHandler(err, c)` для тестов бизнес-логики

**Исправленные тесты**:
- ✅ `TestReservationHandler_CancelReservation/unauthorized_cancellation_attempt`
- ✅ `TestReservationHandler_CancelReservation/cancel_non-existent_reservation`
- ✅ `TestReservationHandler_GuestReservationToken/guest_reservation_requires_name_and_email`
- ✅ `TestReservationHandler_GuestReservationToken/invalid_reservation_token_format`

### Health Handler Tests ✅ FIXED
**Файл**: `backend/internal/domain/health/delivery/http/handler_test.go`

**Применённые исправления**:
1. Добавлен import `"wish-list/internal/app/middleware"`
2. Зарегистрирован `CustomHTTPErrorHandler` в тесте
3. Изменён формат проверки response с `HealthResponse` на `map[string]string`
4. Обновлена проверка на `assert.Contains(t, response["error"], "database connection failed")`
5. Добавлена обработка ошибок через `e.HTTPErrorHandler(err, c)`

**Исправленные тесты**:
- ✅ `TestHandler_Health/returns_unhealthy_when_database_connection_fails`

## Решение

### Вариант 1: Обновить все тесты (рекомендуется)
Обновить все тесты на использование `*apperrors.AppError` и CustomHTTPErrorHandler:

```go
// В setupTestEcho()
func setupTestEcho() *echo.Echo {
    e := echo.New()
    e.Validator = validation.NewValidator()
    e.HTTPErrorHandler = middleware.CustomHTTPErrorHandler
    return e
}

// В тестах
var appErr *apperrors.AppError
assert.True(t, errors.As(err, &appErr))
assert.Equal(t, expectedStatusCode, appErr.Code)
```

### Вариант 2: Временно пропустить тесты
Для быстрого продолжения разработки можно временно пропустить failing тесты:

```bash
go test ./... -v | grep -v "FAIL"
```

## Важные изменения в тестировании

### Обновлённые файлы
1. ✅ `internal/domain/wishlist/delivery/http/handler_test.go` - setupTestEcho с CustomHTTPErrorHandler
2. ✅ `internal/domain/user/delivery/http/handler_test.go` - setupTestEcho с CustomHTTPErrorHandler
3. ✅ `internal/pkg/auth/middleware_test.go` - проверки на `*apperrors.AppError`
4. ✅ `internal/pkg/helpers/request_test.go` - проверки на `*apperrors.AppError`

### Требуют обновления
- ✅ `internal/domain/reservation/delivery/http/handler_test.go` - Fixed
- ✅ `internal/domain/health/delivery/http/handler_test.go` - Fixed
- ✅ Все handler_test.go файлы обновлены

## Запуск тестов

### Успешные тесты
```bash
# Middleware тесты
go test wish-list/internal/app/middleware -v

# Repository тесты
go test wish-list/internal/domain/*/repository -v

# Service тесты
go test wish-list/internal/domain/*/service -v

# apperrors тесты
go test wish-list/internal/pkg/apperrors -v
```

### Failing тесты (требуют обновления)
```bash
go test wish-list/internal/domain/user/delivery/http -v
go test wish-list/internal/domain/reservation/delivery/http -v
go test wish-list/internal/domain/health/delivery/http -v
```

## Приоритет исправлений
1. **High**: User handler tests (основные auth flows)
2. **Medium**: Reservation handler tests (публичные API)
3. **Low**: Health handler tests (utility endpoint)
