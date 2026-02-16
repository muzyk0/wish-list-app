# План по переписыванию тестов на Вариант 2

## Цель
Переписать тесты для проверки error types напрямую вместо обработки через `CustomHTTPErrorHandler`, сделав их настоящими unit tests.

## Принципы нового подхода

### Для валидационных ошибок (уже правильно)
```go
err := handler.SomeMethod(c)

require.Error(t, err)
var appErr *apperrors.AppError
require.True(t, errors.As(err, &appErr))
assert.Equal(t, http.StatusBadRequest, appErr.Code)
```

### Для ошибок бизнес-логики (сейчас неправильно)
```go
// БЫЛО (неправильно):
err := handler.SomeMethod(c)
if err != nil {
    e.HTTPErrorHandler(err, c)
}
assert.Equal(t, http.StatusConflict, rec.Code)

// ДОЛЖНО БЫТЬ (правильно):
err := handler.SomeMethod(c)

require.Error(t, err)
var appErr *apperrors.AppError
require.True(t, errors.As(err, &appErr))
assert.Equal(t, http.StatusConflict, appErr.Code)
assert.Contains(t, appErr.Message, "expected error message")
```

## Этап 1: User Handler Tests

**Файл**: `backend/internal/domain/user/delivery/http/handler_test.go`

### Тесты для рефакторинга:

1. **TestUserHandler_Register_Conflict** (lines 283-327)
   - **Текущий код**: Lines 318-324
     ```go
     if err != nil {
         e.HTTPErrorHandler(err, c)
     }
     assert.Equal(t, nethttp.StatusConflict, rec.Code)
     ```
   - **Новый код**:
     ```go
     require.Error(t, err)
     var appErr *apperrors.AppError
     require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
     assert.Equal(t, nethttp.StatusConflict, appErr.Code)
     assert.Contains(t, appErr.Message, "User with this email already exists")
     ```
   - **Ожидаемое сообщение**: "User with this email already exists"

2. **TestUserHandler_Login_Unauthorized** (lines 329-364)
   - **Текущий код**: Lines 355-361
     ```go
     if err != nil {
         e.HTTPErrorHandler(err, c)
     }
     assert.Equal(t, nethttp.StatusUnauthorized, rec.Code)
     ```
   - **Новый код**:
     ```go
     require.Error(t, err)
     var appErr *apperrors.AppError
     require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
     assert.Equal(t, nethttp.StatusUnauthorized, appErr.Code)
     assert.Contains(t, appErr.Message, "Invalid credentials")
     ```
   - **Ожидаемое сообщение**: "Invalid credentials"

3. **TestUserHandler_GetProfile/unauthenticated_request_returns_unauthorized** (lines 405-419)
   - **Текущий код**: Lines 413-418
     ```go
     if err != nil {
         e.HTTPErrorHandler(err, c)
     }
     assert.Equal(t, nethttp.StatusNotFound, rec.Code)
     ```
   - **Новый код**:
     ```go
     require.Error(t, err)
     var appErr *apperrors.AppError
     require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
     assert.Equal(t, nethttp.StatusNotFound, appErr.Code)
     assert.Contains(t, appErr.Message, "User not found")
     ```
   - **Ожидаемое сообщение**: "User not found"

4. **TestUserHandler_GetProfile/user_not_found_returns_not_found** (lines 421-439)
   - **Текущий код**: Lines 431-436
     ```go
     if err != nil {
         e.HTTPErrorHandler(err, c)
     }
     assert.Equal(t, nethttp.StatusNotFound, rec.Code)
     ```
   - **Новый код**:
     ```go
     require.Error(t, err)
     var appErr *apperrors.AppError
     require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
     assert.Equal(t, nethttp.StatusNotFound, appErr.Code)
     assert.Contains(t, appErr.Message, "User not found")
     ```
   - **Ожидаемое сообщение**: "User not found"

5. **TestUserHandler_GetProfile/other_errors_return_internal_server_error** (lines 441-459)
   - **Текущий код**: Lines 451-456
     ```go
     if err != nil {
         e.HTTPErrorHandler(err, c)
     }
     assert.Equal(t, nethttp.StatusInternalServerError, rec.Code)
     ```
   - **Новый код**:
     ```go
     require.Error(t, err)
     var appErr *apperrors.AppError
     require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
     assert.Equal(t, nethttp.StatusInternalServerError, appErr.Code)
     assert.Contains(t, appErr.Message, "Failed to retrieve user profile")
     ```
   - **Ожидаемое сообщение**: "Failed to retrieve user profile"

6. **TestUserHandler_UpdateProfile/update_profile_unauthorized** (lines 502-530)
   - **Текущий код**: Lines 524-529
     ```go
     if err != nil {
         e.HTTPErrorHandler(err, c)
     }
     assert.Equal(t, nethttp.StatusInternalServerError, rec.Code)
     ```
   - **Новый код**:
     ```go
     require.Error(t, err)
     var appErr *apperrors.AppError
     require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
     assert.Equal(t, nethttp.StatusInternalServerError, appErr.Code)
     ```

7. **TestUserHandler_UpdateProfile/update_profile_service_error** (lines 554-585)
   - **Текущий код**: Lines 576-581
     ```go
     if err != nil {
         e.HTTPErrorHandler(err, c)
     }
     assert.Equal(t, nethttp.StatusInternalServerError, rec.Code)
     ```
   - **Новый код**:
     ```go
     require.Error(t, err)
     var appErr *apperrors.AppError
     require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
     assert.Equal(t, nethttp.StatusInternalServerError, appErr.Code)
     ```

### Паттерн замены:
```go
// Удалить:
if err != nil {
    e.HTTPErrorHandler(err, c)
}
assert.Equal(t, nethttp.StatusXXX, rec.Code)

// Добавить:
require.Error(t, err)
var appErr *apperrors.AppError
require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
assert.Equal(t, nethttp.StatusXXX, appErr.Code)
// Опционально: проверить сообщение
assert.Contains(t, appErr.Message, "expected message")
```

## Этап 2: Reservation Handler Tests

**Файл**: `backend/internal/domain/reservation/delivery/http/handler_test.go`

### Тесты для рефакторинга:

1. **TestReservationHandler_CancelReservation/unauthorized_cancellation_attempt** (lines 221-250)
   - **Текущий код**: Lines 239-249
     ```go
     if err != nil {
         e.HTTPErrorHandler(err, c)
     }
     assert.Equal(t, nethttp.StatusBadRequest, rec.Code)
     var response map[string]string
     err = json.Unmarshal(rec.Body.Bytes(), &response)
     require.NoError(t, err)
     assert.Contains(t, response["error"], "token is required")
     ```
   - **Новый код**:
     ```go
     require.Error(t, err)
     var appErr *apperrors.AppError
     require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
     assert.Equal(t, nethttp.StatusBadRequest, appErr.Code)
     assert.Contains(t, appErr.Message, "Reservation token is required")
     ```
   - **Ожидаемое сообщение**: "Reservation token is required for unauthenticated cancellations"

2. **TestReservationHandler_CancelReservation/cancel_non-existent_reservation** (lines 248-279)
   - **Текущий код**: Lines 271-276
     ```go
     if err != nil {
         e.HTTPErrorHandler(err, c)
     }
     assert.Equal(t, nethttp.StatusInternalServerError, rec.Code)
     ```
   - **Новый код**:
     ```go
     require.Error(t, err)
     var appErr *apperrors.AppError
     require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
     assert.Equal(t, nethttp.StatusInternalServerError, appErr.Code)
     assert.Contains(t, appErr.Message, "Failed to process request")
     ```
   - **Ожидаемое сообщение**: "Failed to process request"

3. **TestReservationHandler_GuestReservationToken/guest_reservation_requires_name_and_email** (lines 350-382)
   - **Текущий код**: Lines 371-381
     ```go
     if err != nil {
         e.HTTPErrorHandler(err, c)
     }
     assert.Equal(t, nethttp.StatusBadRequest, rec.Code)
     var response map[string]string
     err = json.Unmarshal(rec.Body.Bytes(), &response)
     require.NoError(t, err)
     assert.Contains(t, response["error"], "Guest name and email are required")
     ```
   - **Новый код**:
     ```go
     require.Error(t, err)
     var appErr *apperrors.AppError
     require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
     assert.Equal(t, nethttp.StatusBadRequest, appErr.Code)
     assert.Contains(t, appErr.Message, "Guest name and email are required")
     ```
   - **Ожидаемое сообщение**: "Guest name and email are required for unauthenticated reservations"

### Паттерн замены:
```go
// Удалить:
if err != nil {
    e.HTTPErrorHandler(err, c)
}
assert.Equal(t, nethttp.StatusBadRequest, rec.Code)

var response map[string]string
err = json.Unmarshal(rec.Body.Bytes(), &response)
require.NoError(t, err)
assert.Contains(t, response["error"], "expected message")

// Добавить:
require.Error(t, err)
var appErr *apperrors.AppError
require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
assert.Equal(t, nethttp.StatusBadRequest, appErr.Code)
assert.Contains(t, appErr.Message, "expected message")
```

## Этап 3: Health Handler Tests

**Файл**: `backend/internal/domain/health/delivery/http/handler_test.go`

### Тесты для рефакторинга:

1. **TestHandler_Health/returns_unhealthy_when_database_connection_fails** (lines 53-88)
   - **Текущий код**: Lines 75-85
     ```go
     if err != nil {
         e.HTTPErrorHandler(err, c)
     }
     assert.Equal(t, nethttp.StatusServiceUnavailable, rec.Code)
     var response map[string]string
     err = json.Unmarshal(rec.Body.Bytes(), &response)
     require.NoError(t, err)
     assert.Contains(t, response["error"], "database connection failed")
     ```
   - **Новый код**:
     ```go
     require.Error(t, err)
     var appErr *apperrors.AppError
     require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
     assert.Equal(t, nethttp.StatusServiceUnavailable, appErr.Code)
     assert.Contains(t, appErr.Message, "database connection failed")
     ```
   - **Ожидаемое сообщение**: "database connection failed"

### Паттерн замены:
```go
// Удалить:
if err != nil {
    e.HTTPErrorHandler(err, c)
}
assert.Equal(t, nethttp.StatusServiceUnavailable, rec.Code)

var response map[string]string
err = json.Unmarshal(rec.Body.Bytes(), &response)
require.NoError(t, err)
assert.Contains(t, response["error"], "database connection failed")

// Добавить:
require.Error(t, err)
var appErr *apperrors.AppError
require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
assert.Equal(t, nethttp.StatusServiceUnavailable, appErr.Code)
assert.Contains(t, appErr.Message, "database connection failed")
```

## Этап 4: Удаление ненужного кода

После рефакторинга можно удалить:

1. **Регистрацию error handler в некоторых тестах**:
   ```go
   // Можно удалить, если тест не проверяет HTTP response напрямую:
   e.HTTPErrorHandler = middleware.CustomHTTPErrorHandler
   ```

   **Важно**: Оставить регистрацию в `setupTestEcho()` для тестов, которые проверяют успешные случаи (они пишут HTTP response напрямую).

2. **Import middleware** (если больше не используется):
   - В `health/handler_test.go` можно удалить `import "wish-list/internal/app/middleware"`, если он используется только для error handler

3. **Import encoding/json** (в некоторых тестах):
   - Если тест больше не парсит JSON response, можно удалить import

## Этап 5: Добавление дополнительных проверок

Для более строгих тестов можно добавить проверки:

1. **Проверка внутренней ошибки (для Internal Server Error)**:
   ```go
   require.Error(t, err)
   var appErr *apperrors.AppError
   require.True(t, errors.As(err, &appErr))
   assert.Equal(t, nethttp.StatusInternalServerError, appErr.Code)

   // Дополнительно: проверить, что внутренняя ошибка сохранена
   assert.NotNil(t, appErr.Err, "Internal error should be wrapped")
   ```

2. **Проверка конкретных полей ошибки**:
   ```go
   require.Error(t, err)
   var appErr *apperrors.AppError
   require.True(t, errors.As(err, &appErr))
   assert.Equal(t, nethttp.StatusBadRequest, appErr.Code)
   assert.Contains(t, appErr.Message, "expected message")

   // Дополнительно: проверить Details если есть
   if appErr.Details != nil {
       assert.Contains(t, appErr.Details, "field")
   }
   ```

## Этап 6: Проверка

После рефакторинга запустить все тесты:

```bash
# User handler tests
go test wish-list/internal/domain/user/delivery/http -v

# Reservation handler tests
go test wish-list/internal/domain/reservation/delivery/http -v

# Health handler tests
go test wish-list/internal/domain/health/delivery/http -v

# Все вместе
go test ./internal/domain/.../http -v

# С покрытием
go test ./internal/domain/.../http -v -cover
```

## Преимущества после рефакторинга

1. ✅ **Настоящие unit tests** - тестируют только handler logic
2. ✅ **Независимость** - не зависят от error handler implementation
3. ✅ **Простота** - меньше кода, проще понять
4. ✅ **Скорость** - быстрее работают (не вызывают error handler)
5. ✅ **Ясность** - явно видно, какую ошибку ожидаем
6. ✅ **Поддерживаемость** - изменения в error handler не ломают тесты
7. ✅ **Изоляция** - каждый тест проверяет только одну вещь

## Дополнительные улучшения (опционально)

### 1. Создать helper функцию для проверки ошибок:

```go
// В test файле добавить helper:
func assertAppError(t *testing.T, err error, expectedCode int, expectedMessage string) {
    t.Helper()
    require.Error(t, err)
    var appErr *apperrors.AppError
    require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
    assert.Equal(t, expectedCode, appErr.Code)
    if expectedMessage != "" {
        assert.Contains(t, appErr.Message, expectedMessage)
    }
}

// Использование:
err := handler.SomeMethod(c)
assertAppError(t, err, nethttp.StatusBadRequest, "expected message")
```

### 2. Создать таблично-ориентированные тесты:

```go
func TestUserHandler_ErrorCases(t *testing.T) {
    tests := []struct {
        name           string
        setupMock      func(*MockUserService)
        expectedCode   int
        expectedMessage string
    }{
        {
            name: "user not found",
            setupMock: func(m *MockUserService) {
                m.On("GetUser", mock.Anything, mock.Anything).
                    Return(nil, userservice.ErrUserNotFound)
            },
            expectedCode: nethttp.StatusNotFound,
            expectedMessage: "User not found",
        },
        // ... другие случаи
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... setup
            tt.setupMock(mockService)
            err := handler.GetProfile(c)
            assertAppError(t, err, tt.expectedCode, tt.expectedMessage)
        })
    }
}
```

## Чек-лист выполнения

- [ ] Этап 1: Рефакторинг User Handler Tests (7 тестов)
  - [ ] TestUserHandler_Register_Conflict
  - [ ] TestUserHandler_Login_Unauthorized
  - [ ] TestUserHandler_GetProfile/unauthenticated_request_returns_unauthorized
  - [ ] TestUserHandler_GetProfile/user_not_found_returns_not_found
  - [ ] TestUserHandler_GetProfile/other_errors_return_internal_server_error
  - [ ] TestUserHandler_UpdateProfile/update_profile_unauthorized
  - [ ] TestUserHandler_UpdateProfile/update_profile_service_error

- [ ] Этап 2: Рефакторинг Reservation Handler Tests (3 теста)
  - [ ] TestReservationHandler_CancelReservation/unauthorized_cancellation_attempt
  - [ ] TestReservationHandler_CancelReservation/cancel_non-existent_reservation
  - [ ] TestReservationHandler_GuestReservationToken/guest_reservation_requires_name_and_email

- [ ] Этап 3: Рефакторинг Health Handler Tests (1 тест)
  - [ ] TestHandler_Health/returns_unhealthy_when_database_connection_fails

- [ ] Этап 4: Удаление ненужного кода
  - [ ] Проверить использование middleware.CustomHTTPErrorHandler
  - [ ] Удалить ненужные imports (json, middleware)

- [ ] Этап 5: Добавление дополнительных проверок (опционально)
  - [ ] Проверка внутренних ошибок
  - [ ] Проверка Details полей

- [ ] Этап 6: Проверка всех тестов
  - [ ] go test user handler
  - [ ] go test reservation handler
  - [ ] go test health handler
  - [ ] go test с покрытием

- [ ] Этап 7: Обновление документации
  - [ ] Обновить TEST-UPDATES-NEEDED.md
  - [ ] Добавить комментарии в код о новом подходе

- [ ] Опционально: Создать helper функции
- [ ] Опционально: Преобразовать в таблично-ориентированные тесты

---

**Общее количество тестов для рефакторинга**: 11 тестов

**Ожидаемое время**: 30-45 минут

**Приоритет**: Medium (улучшение качества тестов, не критично для функциональности)
