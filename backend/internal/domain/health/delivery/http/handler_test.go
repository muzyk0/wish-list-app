package http

import (
	"database/sql"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"wish-list/internal/app/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Health(t *testing.T) {
	t.Run("returns healthy when database is connected", func(t *testing.T) {
		// Create mock database
		mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
		dbWrapper := &database.DB{DB: sqlxDB}

		handler := NewHandler(dbWrapper)

		// Expect ping to succeed
		mock.ExpectPing()

		e := echo.New()
		req := httptest.NewRequest(nethttp.MethodGet, "/health", nethttp.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.Health(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusOK, rec.Code)

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
		dbWrapper := &database.DB{DB: sqlxDB}

		handler := NewHandler(dbWrapper)

		// Expect ping to fail
		mock.ExpectPing().WillReturnError(sql.ErrConnDone)

		e := echo.New()
		req := httptest.NewRequest(nethttp.MethodGet, "/health", nethttp.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.Health(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusServiceUnavailable, rec.Code)

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
		dbWrapper := &database.DB{DB: sqlxDB}

		handler := NewHandler(dbWrapper)

		// Expect ping
		mock.ExpectPing()

		e := echo.New()
		req := httptest.NewRequest(nethttp.MethodGet, "/health", nethttp.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.Health(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusOK, rec.Code)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
