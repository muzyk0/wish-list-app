package helpers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedPage   int
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:           "default values when no params provided",
			queryParams:    map[string]string{},
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "valid page and limit",
			queryParams: map[string]string{
				"page":  "2",
				"limit": "20",
			},
			expectedPage:   2,
			expectedLimit:  20,
			expectedOffset: 20,
		},
		{
			name: "page only",
			queryParams: map[string]string{
				"page": "3",
			},
			expectedPage:   3,
			expectedLimit:  10,
			expectedOffset: 20,
		},
		{
			name: "limit only",
			queryParams: map[string]string{
				"limit": "50",
			},
			expectedPage:   1,
			expectedLimit:  50,
			expectedOffset: 0,
		},
		{
			name: "zero page defaults to 1",
			queryParams: map[string]string{
				"page": "0",
			},
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "negative page defaults to 1",
			queryParams: map[string]string{
				"page": "-5",
			},
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "zero limit defaults to 10",
			queryParams: map[string]string{
				"limit": "0",
			},
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "negative limit defaults to 10",
			queryParams: map[string]string{
				"limit": "-10",
			},
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "limit at max boundary (100)",
			queryParams: map[string]string{
				"limit": "100",
			},
			expectedPage:   1,
			expectedLimit:  100,
			expectedOffset: 0,
		},
		{
			name: "limit over max (101) defaults to 10",
			queryParams: map[string]string{
				"limit": "101",
			},
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "limit way over max (1000) defaults to 10",
			queryParams: map[string]string{
				"limit": "1000",
			},
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "invalid page string defaults to 1",
			queryParams: map[string]string{
				"page": "abc",
			},
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "invalid limit string defaults to 10",
			queryParams: map[string]string{
				"limit": "xyz",
			},
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "decimal page number defaults to 1",
			queryParams: map[string]string{
				"page": "1.5",
			},
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "decimal limit number defaults to 10",
			queryParams: map[string]string{
				"limit": "10.5",
			},
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "correct offset calculation for page 5 with limit 25",
			queryParams: map[string]string{
				"page":  "5",
				"limit": "25",
			},
			expectedPage:   5,
			expectedLimit:  25,
			expectedOffset: 100,
		},
		{
			name: "correct offset calculation for page 10 with limit 50",
			queryParams: map[string]string{
				"page":  "10",
				"limit": "50",
			},
			expectedPage:   10,
			expectedLimit:  50,
			expectedOffset: 450,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Echo context with query params
			e := echo.New()
			q := make(url.Values)
			for key, value := range tt.queryParams {
				q.Set(key, value)
			}
			req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Parse pagination
			result := ParsePagination(c)

			// Assert results
			assert.Equal(t, tt.expectedPage, result.Page, "Page mismatch")
			assert.Equal(t, tt.expectedLimit, result.Limit, "Limit mismatch")
			assert.Equal(t, tt.expectedOffset, result.Offset, "Offset mismatch")
		})
	}
}

func TestParsePaginationEdgeCases(t *testing.T) {
	t.Run("empty query string", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		result := ParsePagination(c)

		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 10, result.Limit)
		assert.Equal(t, 0, result.Offset)
	})

	t.Run("multiple query params with same key (uses first)", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/?page=1&page=3", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		result := ParsePagination(c)

		// Echo uses the first value when multiple params with same key exist
		assert.Equal(t, 1, result.Page)
	})

	t.Run("very large page number", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/?page=999999", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		result := ParsePagination(c)

		assert.Equal(t, 999999, result.Page)
		assert.Equal(t, 9999980, result.Offset)
	})
}
