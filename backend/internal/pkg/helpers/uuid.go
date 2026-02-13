package helpers

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// ParseUUID parses a string into pgtype.UUID.
// Returns error response if parsing fails.
//
// Example usage in handler:
//
//	func (h *Handler) CreateReservation(c echo.Context) error {
//	    userIDStr, _, _, _ := auth.GetUserFromContext(c)
//	    userID, err := helpers.ParseUUID(c, userIDStr)
//	    if err != nil {
//	        return err
//	    }
//	    // userID is now a valid pgtype.UUID
//	}
func ParseUUID(c echo.Context, uuidStr string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(uuidStr); err != nil {
		return uuid, c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid UUID format",
		})
	}
	return uuid, nil
}

// MustParseUUID parses a string into pgtype.UUID without returning HTTP error.
// Useful when you need just the UUID value and will handle errors separately.
// Returns invalid UUID (Valid=false) if parsing fails.
//
// Example usage:
//
//	userID := helpers.MustParseUUID(userIDStr)
//	if !userID.Valid {
//	    // Handle error
//	}
func MustParseUUID(uuidStr string) pgtype.UUID {
	var uuid pgtype.UUID
	_ = uuid.Scan(uuidStr)
	return uuid
}
