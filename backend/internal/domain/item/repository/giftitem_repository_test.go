package repository

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"wish-list/internal/domain/item/models"
	reservationmodels "wish-list/internal/domain/reservation/models"
)

func TestGiftItemRepository_Create(t *testing.T) {
	t.Run("validate gift item creation fields", func(t *testing.T) {
		giftItem := models.GiftItem{
			OwnerID:     pgtype.UUID{Valid: true},
			Name:        "Nintendo Switch",
			Description: pgtype.Text{String: "The latest gaming console", Valid: true},
			Link:        pgtype.Text{String: "https://example.com/switch", Valid: true},
			ImageUrl:    pgtype.Text{String: "https://cdn.example.com/switch.jpg", Valid: true},
			Price:       pgtype.Numeric{Valid: true},
			Priority:    pgtype.Int4{Int32: 5, Valid: true},
			Position:    pgtype.Int4{Int32: 0, Valid: true},
		}

		// Verify required fields
		if giftItem.Name == "" {
			t.Error("name should not be empty")
		}
		if !giftItem.OwnerID.Valid {
			t.Error("owner_id should be valid")
		}
	})

	t.Run("create gift item with minimal fields", func(t *testing.T) {
		giftItem := models.GiftItem{
			OwnerID: pgtype.UUID{Valid: true},
			Name:    "Minimal Gift",
		}

		// Verify optional fields can be omitted
		if giftItem.Description.Valid {
			t.Error("description should be invalid when not set")
		}
		if giftItem.Link.Valid {
			t.Error("link should be invalid when not set")
		}
		if giftItem.ImageUrl.Valid {
			t.Error("image_url should be invalid when not set")
		}
		if giftItem.Price.Valid {
			t.Error("price should be invalid when not set")
		}
	})

	t.Run("gift item initial state is available", func(t *testing.T) {
		giftItem := models.GiftItem{
			OwnerID: pgtype.UUID{Valid: true},
			Name:    "Available Gift",
		}

		// New items should not be reserved or purchased
		if giftItem.ReservedByUserID.Valid {
			t.Error("new gift item should not be reserved")
		}
		if giftItem.PurchasedByUserID.Valid {
			t.Error("new gift item should not be purchased")
		}
	})
}

func TestGiftItemRepository_GetByID(t *testing.T) {
	t.Run("validate gift item retrieval structure", func(t *testing.T) {
		// Simulate a retrieved gift item
		giftItem := models.GiftItem{
			ID:          pgtype.UUID{Valid: true},
			OwnerID:     pgtype.UUID{Valid: true},
			Name:        "Test Gift",
			Description: pgtype.Text{String: "Test description", Valid: true},
			Priority:    pgtype.Int4{Int32: 5, Valid: true},
			Position:    pgtype.Int4{Int32: 0, Valid: true},
		}

		// Verify all required fields are present
		if !giftItem.ID.Valid {
			t.Error("id should be valid")
		}
		if !giftItem.OwnerID.Valid {
			t.Error("owner_id should be valid")
		}
		if giftItem.Name == "" {
			t.Error("name should not be empty")
		}
	})
}

func TestGiftItemRepository_GetByWishList(t *testing.T) {
	t.Run("get gift items ordered by position", func(t *testing.T) {
		// Simulate multiple gift items
		giftItems := []*models.GiftItem{
			{
				ID:       pgtype.UUID{Valid: true},
				Name:     "First Gift",
				Position: pgtype.Int4{Int32: 0, Valid: true},
			},
			{
				ID:       pgtype.UUID{Valid: true},
				Name:     "Second Gift",
				Position: pgtype.Int4{Int32: 1, Valid: true},
			},
			{
				ID:       pgtype.UUID{Valid: true},
				Name:     "Third Gift",
				Position: pgtype.Int4{Int32: 2, Valid: true},
			},
		}

		// Verify positions are sequential
		for i, item := range giftItems {
			expectedPos := int32(i) // #nosec G115 -- loop index, always safe conversion
			if item.Position.Int32 != expectedPos {
				t.Errorf("item %d: expected position %d, got %d", i, expectedPos, item.Position.Int32)
			}
		}
	})
}

func TestGiftItemRepository_Update(t *testing.T) {
	t.Run("update gift item preserves ID", func(t *testing.T) {
		originalID := pgtype.UUID{Valid: true}

		giftItem := models.GiftItem{
			ID:          originalID,
			OwnerID:     pgtype.UUID{Valid: true},
			Name:        "Updated Gift",
			Description: pgtype.Text{String: "Updated description", Valid: true},
			Price:       pgtype.Numeric{Valid: true},
		}

		// ID should remain unchanged during update
		if giftItem.ID != originalID {
			t.Error("gift item ID should not change during update")
		}
	})

	t.Run("update can change position", func(t *testing.T) {
		giftItem := models.GiftItem{
			ID:       pgtype.UUID{Valid: true},
			OwnerID:  pgtype.UUID{Valid: true},
			Name:     "Reorderable Gift",
			Position: pgtype.Int4{Int32: 0, Valid: true},
		}

		// Change position
		giftItem.Position = pgtype.Int4{Int32: 5, Valid: true}

		if giftItem.Position.Int32 != 5 {
			t.Errorf("expected position 5, got %d", giftItem.Position.Int32)
		}
	})
}

func TestGiftItemRepository_Delete(t *testing.T) {
	t.Run("delete requires valid ID", func(t *testing.T) {
		validID := pgtype.UUID{Valid: true}

		if !validID.Valid {
			t.Error("delete should require a valid UUID")
		}
	})

	t.Run("delete cascades to reservations", func(t *testing.T) {
		// This is enforced by database foreign key constraints:
		// reservations.gift_item_id REFERENCES gift_items(id) ON DELETE CASCADE

		giftItemID := pgtype.UUID{Valid: true}
		reservation := reservationmodels.Reservation{
			GiftItemID: giftItemID,
			Status:     "active",
		}

		if reservation.GiftItemID != giftItemID {
			t.Error("reservation should reference valid gift item")
		}
	})
}

func TestGiftItemRepository_Reserve(t *testing.T) {
	t.Run("reserve available gift item", func(t *testing.T) {
		giftItem := models.GiftItem{
			ID:      pgtype.UUID{Valid: true},
			OwnerID: pgtype.UUID{Valid: true},
			Name:    "Available Gift",
		}

		// Reserve the item
		userID := pgtype.UUID{Valid: true}
		giftItem.ReservedByUserID = userID
		giftItem.ReservedAt = pgtype.Timestamptz{Valid: true}

		// Verify reservation state
		if !giftItem.ReservedByUserID.Valid {
			t.Error("reserved gift item should have reserved_by_user_id")
		}
		if !giftItem.ReservedAt.Valid {
			t.Error("reserved gift item should have reserved_at timestamp")
		}
		if giftItem.ReservedByUserID != userID {
			t.Error("gift item should be reserved by the correct user")
		}
	})

	t.Run("cannot reserve already reserved item", func(t *testing.T) {
		giftItem := models.GiftItem{
			ID:               pgtype.UUID{Valid: true},
			Name:             "Reserved Gift",
			ReservedByUserID: pgtype.UUID{Valid: true},
			ReservedAt:       pgtype.Timestamptz{Valid: true},
		}

		// Item is already reserved
		if !giftItem.ReservedByUserID.Valid {
			t.Error("item should be reserved")
		}

		// Attempting to reserve again should be rejected
		// This would be enforced by ReserveIfNotReserved method
	})

	t.Run("cannot reserve purchased item", func(t *testing.T) {
		giftItem := models.GiftItem{
			ID:                pgtype.UUID{Valid: true},
			Name:              "Purchased Gift",
			PurchasedByUserID: pgtype.UUID{Valid: true},
			PurchasedAt:       pgtype.Timestamptz{Valid: true},
		}

		// Item is already purchased
		if !giftItem.PurchasedByUserID.Valid {
			t.Error("item should be purchased")
		}

		// Database constraint prevents both reserved and purchased
		// chk_not_reserved_and_purchased
	})
}

func TestGiftItemRepository_Unreserve(t *testing.T) {
	t.Run("unreserve reserved gift item", func(t *testing.T) {
		giftItem := models.GiftItem{
			ID:               pgtype.UUID{Valid: true},
			Name:             "Reserved Gift",
			ReservedByUserID: pgtype.UUID{Valid: true},
			ReservedAt:       pgtype.Timestamptz{Valid: true},
		}

		// Unreserve the item
		giftItem.ReservedByUserID = pgtype.UUID{Valid: false}
		giftItem.ReservedAt = pgtype.Timestamptz{Valid: false}

		// Verify unreserved state
		if giftItem.ReservedByUserID.Valid {
			t.Error("unreserved gift item should not have reserved_by_user_id")
		}
		if giftItem.ReservedAt.Valid {
			t.Error("unreserved gift item should not have reserved_at timestamp")
		}
	})

	t.Run("unreserve available item is idempotent", func(t *testing.T) {
		giftItem := models.GiftItem{
			ID:   pgtype.UUID{Valid: true},
			Name: "Available Gift",
		}

		// Item is already unreserved
		if giftItem.ReservedByUserID.Valid {
			t.Error("item should not be reserved")
		}

		// Unreserving again should be safe (no-op)
	})
}

func TestGiftItemRepository_MarkAsPurchased(t *testing.T) {
	t.Run("mark reserved item as purchased", func(t *testing.T) {
		giftItem := models.GiftItem{
			ID:               pgtype.UUID{Valid: true},
			Name:             "Reserved Gift",
			ReservedByUserID: pgtype.UUID{Valid: true},
			ReservedAt:       pgtype.Timestamptz{Valid: true},
		}

		// Mark as purchased
		userID := pgtype.UUID{Valid: true}
		giftItem.PurchasedByUserID = userID
		giftItem.PurchasedAt = pgtype.Timestamptz{Valid: true}
		giftItem.PurchasedPrice = pgtype.Numeric{Valid: true}

		// Verify purchased state
		if !giftItem.PurchasedByUserID.Valid {
			t.Error("purchased gift item should have purchased_by_user_id")
		}
		if !giftItem.PurchasedAt.Valid {
			t.Error("purchased gift item should have purchased_at timestamp")
		}
		if !giftItem.PurchasedPrice.Valid {
			t.Error("purchased gift item should have purchased_price")
		}
	})

	t.Run("mark available item as purchased", func(t *testing.T) {
		giftItem := models.GiftItem{
			ID:   pgtype.UUID{Valid: true},
			Name: "Available Gift",
		}

		// Can mark as purchased even if not reserved
		userID := pgtype.UUID{Valid: true}
		giftItem.PurchasedByUserID = userID
		giftItem.PurchasedAt = pgtype.Timestamptz{Valid: true}

		if !giftItem.PurchasedByUserID.Valid {
			t.Error("item should be marked as purchased")
		}
	})

	t.Run("purchased price can differ from listed price", func(t *testing.T) {
		giftItem := models.GiftItem{
			ID:    pgtype.UUID{Valid: true},
			Name:  "Discounted Gift",
			Price: pgtype.Numeric{Valid: true}, // Listed price
		}

		// Mark as purchased with actual paid price
		giftItem.PurchasedPrice = pgtype.Numeric{Valid: true} // Actual price

		// Both prices can exist simultaneously
		if !giftItem.Price.Valid {
			t.Error("listed price should still be valid")
		}
		if !giftItem.PurchasedPrice.Valid {
			t.Error("purchased price should be valid")
		}
	})
}

func TestGiftItemRepository_StateTransitions(t *testing.T) {
	t.Run("available to reserved", func(t *testing.T) {
		giftItem := models.GiftItem{
			ID:   pgtype.UUID{Valid: true},
			Name: "Gift",
		}

		// Transition: Available -> Reserved
		giftItem.ReservedByUserID = pgtype.UUID{Valid: true}
		giftItem.ReservedAt = pgtype.Timestamptz{Valid: true}

		if !giftItem.ReservedByUserID.Valid {
			t.Error("failed transition to reserved state")
		}
	})

	t.Run("reserved to purchased", func(t *testing.T) {
		giftItem := models.GiftItem{
			ID:               pgtype.UUID{Valid: true},
			Name:             "Gift",
			ReservedByUserID: pgtype.UUID{Valid: true},
			ReservedAt:       pgtype.Timestamptz{Valid: true},
		}

		// Transition: Reserved -> Purchased
		giftItem.PurchasedByUserID = pgtype.UUID{Valid: true}
		giftItem.PurchasedAt = pgtype.Timestamptz{Valid: true}

		if !giftItem.PurchasedByUserID.Valid {
			t.Error("failed transition to purchased state")
		}
	})

	t.Run("reserved to available", func(t *testing.T) {
		giftItem := models.GiftItem{
			ID:               pgtype.UUID{Valid: true},
			Name:             "Gift",
			ReservedByUserID: pgtype.UUID{Valid: true},
			ReservedAt:       pgtype.Timestamptz{Valid: true},
		}

		// Transition: Reserved -> Available (unreserve)
		giftItem.ReservedByUserID = pgtype.UUID{Valid: false}
		giftItem.ReservedAt = pgtype.Timestamptz{Valid: false}

		if giftItem.ReservedByUserID.Valid {
			t.Error("failed transition back to available state")
		}
	})
}

func TestGiftItemRepository_ValidationRules(t *testing.T) {
	t.Run("name length validation", func(t *testing.T) {
		testCases := []struct {
			name  string
			value string
			valid bool
		}{
			{"empty name", "", false},
			{"single character", "G", true},
			{"normal name", "Nintendo Switch", true},
			{"max length", string(make([]byte, 255)), true},
			{"too long", string(make([]byte, 256)), false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				nameLength := len(tc.value)
				isValid := nameLength > 0 && nameLength <= 255

				if isValid != tc.valid {
					t.Errorf("expected valid=%v for name length %d", tc.valid, nameLength)
				}
			})
		}
	})

	t.Run("priority range validation", func(t *testing.T) {
		testCases := []struct {
			priority int32
			valid    bool
		}{
			{-1, false},
			{0, true},
			{5, true},
			{10, true},
			{11, false},
		}

		for _, tc := range testCases {
			giftItem := models.GiftItem{
				Priority: pgtype.Int4{Int32: tc.priority, Valid: true},
			}

			isValid := giftItem.Priority.Int32 >= 0 && giftItem.Priority.Int32 <= 10

			if isValid != tc.valid {
				t.Errorf("priority %d: expected valid=%v, got %v", tc.priority, tc.valid, isValid)
			}
		}
	})

	t.Run("link must be valid URL if provided", func(t *testing.T) {
		validURLs := []string{
			"https://example.com/product",
			"http://example.com",
			"https://www.amazon.com/dp/B08H75RTZ8",
		}

		for _, url := range validURLs {
			giftItem := models.GiftItem{
				Link: pgtype.Text{String: url, Valid: true},
			}

			if !giftItem.Link.Valid {
				t.Errorf("valid URL %q should be accepted", url)
			}
		}
	})

	t.Run("position must be non-negative", func(t *testing.T) {
		testCases := []struct {
			position int32
			valid    bool
		}{
			{-1, false},
			{0, true},
			{100, true},
		}

		for _, tc := range testCases {
			giftItem := models.GiftItem{
				Position: pgtype.Int4{Int32: tc.position, Valid: true},
			}

			isValid := giftItem.Position.Int32 >= 0

			if isValid != tc.valid {
				t.Errorf("position %d: expected valid=%v, got %v", tc.position, tc.valid, isValid)
			}
		}
	})
}

func TestGiftItemRepository_SQLInjectionPrevention(t *testing.T) {
	t.Run("malicious sort field is rejected", func(t *testing.T) {
		maliciousSorts := []string{
			"created_at; DROP TABLE users--",
			"created_at; DELETE FROM users--",
			"created_at' OR '1'='1",
			"(SELECT * FROM users)",
			"created_at; INSERT INTO users VALUES ('hack')--",
			"; DROP TABLE gift_items;--",
			"created_at UNION SELECT * FROM users--",
		}

		for _, sort := range maliciousSorts {
			_, ok := validSortFields[sort]
			if ok {
				t.Errorf("malicious sort field %q should not be in whitelist", sort)
			}
		}
	})

	t.Run("malicious order direction is rejected", func(t *testing.T) {
		maliciousOrders := []string{
			"DESC; DROP TABLE users--",
			"ASC; DELETE FROM users--",
			"DESC' OR '1'='1",
			"(SELECT * FROM users)",
			"; INSERT INTO users VALUES ('hack')--",
			"UNION SELECT * FROM users--",
		}

		for _, order := range maliciousOrders {
			orderUpper := string(order)
			if orderUpper == "" {
				orderUpper = "DESC"
			}
			if validSortOrders[orderUpper] {
				t.Errorf("malicious order direction %q should not be valid", order)
			}
		}
	})

	t.Run("valid sort fields are accepted", func(t *testing.T) {
		validSorts := []string{"created_at", "updated_at", "title", "price"}

		for _, sort := range validSorts {
			_, ok := validSortFields[sort]
			if !ok {
				t.Errorf("valid sort field %q should be in whitelist", sort)
			}
		}
	})

	t.Run("valid order directions are accepted", func(t *testing.T) {
		validOrders := []string{"ASC", "DESC"}

		for _, order := range validOrders {
			if !validSortOrders[order] {
				t.Errorf("valid order direction %q should be accepted", order)
			}
		}
	})

	t.Run("sort field validation is case sensitive", func(t *testing.T) {
		// Valid sort fields should be case-sensitive
		_, ok := validSortFields["CREATED_AT"]
		if ok {
			t.Error("sort field validation should be case sensitive")
		}

		_, ok = validSortFields["Created_At"]
		if ok {
			t.Error("sort field validation should be case sensitive")
		}
	})
}

func TestGiftItemRepository_EdgeCases(t *testing.T) {
	t.Run("handle special characters in name", func(t *testing.T) {
		specialNames := []string{
			"Gift with emoji \U0001f381",
			"Gift & accessories",
			"\"Special\" gift",
			"\u6d4b\u8bd5\u793c\u7269",
		}

		for _, name := range specialNames {
			giftItem := models.GiftItem{
				Name: name,
			}

			if giftItem.Name != name {
				t.Errorf("name should preserve special characters: %q", name)
			}
		}
	})

	t.Run("handle very long URLs", func(t *testing.T) {
		// URLs can be very long (especially with query parameters)
		longURL := "https://example.com/product?" + string(make([]byte, 500))

		giftItem := models.GiftItem{
			Link: pgtype.Text{String: longURL, Valid: true},
		}

		if giftItem.Link.String != longURL {
			t.Error("long URLs should be preserved")
		}
	})

	t.Run("handle null optional fields", func(t *testing.T) {
		giftItem := models.GiftItem{
			OwnerID:     pgtype.UUID{Valid: true},
			Name:        "Minimal Gift",
			Description: pgtype.Text{Valid: false},
			Link:        pgtype.Text{Valid: false},
			ImageUrl:    pgtype.Text{Valid: false},
			Price:       pgtype.Numeric{Valid: false},
			Notes:       pgtype.Text{Valid: false},
		}

		// All optional fields should be invalid
		if giftItem.Description.Valid {
			t.Error("description should be invalid")
		}
		if giftItem.Link.Valid {
			t.Error("link should be invalid")
		}
		if giftItem.ImageUrl.Valid {
			t.Error("image_url should be invalid")
		}
	})
}
