package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	db "wish-list/internal/db/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_Health(t *testing.T) {
	t.Run("returns healthy when database is connected", func(t *testing.T) {
		// Create mock database
		mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		dbWrapper := &db.DB{DB: sqlxDB}

		handler := NewHealthHandler(dbWrapper)

		// Expect ping to succeed
		mock.ExpectPing()

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.Health(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response HealthResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "healthy", response.Status)
		assert.Equal(t, "ok", response.Checks["database"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns unhealthy when database connection fails", func(t *testing.T) {
		// Create mock database
		mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		dbWrapper := &db.DB{DB: sqlxDB}

		handler := NewHealthHandler(dbWrapper)

		// Expect ping to fail
		mock.ExpectPing().WillReturnError(sql.ErrConnDone)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.Health(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

		var response HealthResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unhealthy", response.Status)
		assert.Equal(t, "database connection failed", response.Error)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handles context properly", func(t *testing.T) {
		// Create mock database
		mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		dbWrapper := &db.DB{DB: sqlxDB}

		handler := NewHealthHandler(dbWrapper)

		// Expect ping
		mock.ExpectPing()

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.Health(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

