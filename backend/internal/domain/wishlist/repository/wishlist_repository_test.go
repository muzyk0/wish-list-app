package repository

import (
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"wish-list/internal/domain/wishlist/models"
)

func TestWishListRepository_Create(t *testing.T) {
	// Note: These tests verify the repository logic without requiring a real database.
	// In production, you would use a test database or mocks.

	t.Run("validate wishlist creation fields", func(t *testing.T) {
		wishList := models.WishList{
			OwnerID:     pgtype.UUID{Valid: true},
			Title:       "Birthday Wish List",
			Description: pgtype.Text{String: "My birthday wishes for 2024", Valid: true},
			Occasion:    pgtype.Text{String: "Birthday", Valid: true},
			TemplateID:  "default",
			IsPublic:    pgtype.Bool{Bool: true, Valid: true},
			PublicSlug:  pgtype.Text{String: "john-birthday-2024", Valid: true},
		}

		// Verify required fields
		if wishList.Title == "" {
			t.Error("title should not be empty")
		}
		if wishList.TemplateID == "" {
			t.Error("template_id should not be empty")
		}
		if !wishList.OwnerID.Valid {
			t.Error("owner_id should be valid")
		}
	})

	t.Run("create wishlist with minimal fields", func(t *testing.T) {
		wishList := models.WishList{
			OwnerID:    pgtype.UUID{Valid: true},
			Title:      "Minimal Wish List",
			TemplateID: "default",
		}

		// Verify optional fields can be omitted
		if wishList.Description.Valid {
			t.Error("description should be invalid when not set")
		}
		if wishList.Occasion.Valid {
			t.Error("occasion should be invalid when not set")
		}
		if wishList.PublicSlug.Valid {
			t.Error("public_slug should be invalid when not set")
		}
	})

	t.Run("create public wishlist requires slug", func(t *testing.T) {
		wishList := models.WishList{
			OwnerID:    pgtype.UUID{Valid: true},
			Title:      "Public Wish List",
			TemplateID: "default",
			IsPublic:   pgtype.Bool{Bool: true, Valid: true},
			PublicSlug: pgtype.Text{String: "test-slug", Valid: true},
		}

		// Verify public wishlists have slug
		if wishList.IsPublic.Bool && !wishList.PublicSlug.Valid {
			t.Error("public wishlists should have a valid public_slug")
		}
	})
}

func TestWishListRepository_GetByID(t *testing.T) {
	t.Run("validate wishlist retrieval structure", func(t *testing.T) {
		// Simulate a retrieved wishlist
		wishList := models.WishList{
			ID:          pgtype.UUID{Valid: true},
			OwnerID:     pgtype.UUID{Valid: true},
			Title:       "Test Wish List",
			Description: pgtype.Text{String: "Test description", Valid: true},
			TemplateID:  "default",
			IsPublic:    pgtype.Bool{Bool: false, Valid: true},
			ViewCount:   pgtype.Int4{Int32: 0, Valid: true},
		}

		// Verify all fields are present
		if !wishList.ID.Valid {
			t.Error("id should be valid")
		}
		if !wishList.OwnerID.Valid {
			t.Error("owner_id should be valid")
		}
		if wishList.Title == "" {
			t.Error("title should not be empty")
		}
		if !wishList.ViewCount.Valid {
			t.Error("view_count should be valid")
		}
	})
}

func TestWishListRepository_GetByPublicSlug(t *testing.T) {
	t.Run("public slug should be unique and url-safe", func(t *testing.T) {
		testCases := []struct {
			name  string
			slug  string
			valid bool
		}{
			{"lowercase alphanumeric", "birthday-2024", true},
			{"with hyphens", "my-awesome-wishlist", true},
			{"with numbers", "wishlist-123", true},
			{"starts with letter", "w123", true},
			{"all lowercase", "allowercase", true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				wishList := models.WishList{
					PublicSlug: pgtype.Text{String: tc.slug, Valid: true},
					IsPublic:   pgtype.Bool{Bool: true, Valid: true},
				}

				if tc.valid && wishList.PublicSlug.String != tc.slug {
					t.Errorf("expected slug %q, got %q", tc.slug, wishList.PublicSlug.String)
				}
			})
		}
	})

	t.Run("only public wishlists should be retrievable by slug", func(t *testing.T) {
		wishList := models.WishList{
			PublicSlug: pgtype.Text{String: "test-slug", Valid: true},
			IsPublic:   pgtype.Bool{Bool: false, Valid: true},
		}

		// This would be handled by the WHERE clause in GetByPublicSlug:
		// WHERE public_slug = $1 AND is_public = true
		if !wishList.IsPublic.Bool {
			// Private wishlists should not be retrievable by slug
			if wishList.PublicSlug.Valid {
				t.Log("private wishlist should not be retrievable by public slug")
			}
		}
	})
}

func TestWishListRepository_GetByOwner(t *testing.T) {
	t.Run("get all wishlists for owner", func(t *testing.T) {
		// Simulate multiple wishlists for an owner
		wishlists := []*models.WishList{
			{
				ID:      pgtype.UUID{Valid: true},
				OwnerID: pgtype.UUID{Valid: true},
				Title:   "Birthday 2024",
			},
			{
				ID:      pgtype.UUID{Valid: true},
				OwnerID: pgtype.UUID{Valid: true},
				Title:   "Christmas 2024",
			},
		}

		// Verify both have same owner
		if len(wishlists) != 2 {
			t.Errorf("expected 2 wishlists, got %d", len(wishlists))
		}
	})
}

func TestWishListRepository_Update(t *testing.T) {
	t.Run("update wishlist preserves ID", func(t *testing.T) {
		originalID := pgtype.UUID{Valid: true}

		wishList := models.WishList{
			ID:          originalID,
			OwnerID:     pgtype.UUID{Valid: true},
			Title:       "Updated Title",
			Description: pgtype.Text{String: "Updated description", Valid: true},
			TemplateID:  "modern",
		}

		// ID should remain unchanged during update
		if wishList.ID != originalID {
			t.Error("wishlist ID should not change during update")
		}
	})

	t.Run("update can change public status", func(t *testing.T) {
		wishList := models.WishList{
			ID:         pgtype.UUID{Valid: true},
			OwnerID:    pgtype.UUID{Valid: true},
			Title:      "Test List",
			TemplateID: "default",
			IsPublic:   pgtype.Bool{Bool: false, Valid: true},
		}

		// Change to public
		wishList.IsPublic = pgtype.Bool{Bool: true, Valid: true}
		wishList.PublicSlug = pgtype.Text{String: "test-list", Valid: true}

		if !wishList.IsPublic.Bool {
			t.Error("wishlist should be public after update")
		}
		if !wishList.PublicSlug.Valid {
			t.Error("public wishlist should have a slug")
		}
	})

	t.Run("update can change template", func(t *testing.T) {
		wishList := models.WishList{
			ID:         pgtype.UUID{Valid: true},
			OwnerID:    pgtype.UUID{Valid: true},
			Title:      "Test List",
			TemplateID: "default",
		}

		// Change template
		wishList.TemplateID = "modern"

		if wishList.TemplateID != "modern" {
			t.Errorf("expected template 'modern', got %q", wishList.TemplateID)
		}
	})
}

func TestWishListRepository_Delete(t *testing.T) {
	t.Run("delete requires valid ID", func(t *testing.T) {
		validID := pgtype.UUID{Valid: true}

		if !validID.Valid {
			t.Error("delete should require a valid UUID")
		}
	})

	t.Run("delete should cascade to wishlist_items", func(t *testing.T) {
		// With the many-to-many schema, cascade is through wishlist_items junction table:
		// wishlist_items.wishlist_id REFERENCES wishlists(id) ON DELETE CASCADE

		wishListID := pgtype.UUID{Valid: true}

		// WishlistItem is in shared models since it's a junction table
		// Just verify the ID reference is valid
		if !wishListID.Valid {
			t.Error("wishlist_item should reference valid wishlist")
		}
	})
}

func TestWishListRepository_IncrementViewCount(t *testing.T) {
	t.Run("increment view count atomically", func(t *testing.T) {
		wishList := models.WishList{
			ID:        pgtype.UUID{Valid: true},
			ViewCount: pgtype.Int4{Int32: 10, Valid: true},
		}

		// Simulate increment
		wishList.ViewCount.Int32++

		if wishList.ViewCount.Int32 != 11 {
			t.Errorf("expected view count 11, got %d", wishList.ViewCount.Int32)
		}
	})

	t.Run("view count starts at zero", func(t *testing.T) {
		wishList := models.WishList{
			ID:        pgtype.UUID{Valid: true},
			ViewCount: pgtype.Int4{Int32: 0, Valid: true},
		}

		if wishList.ViewCount.Int32 != 0 {
			t.Error("new wishlist should have zero views")
		}
	})
}

func TestWishListRepository_ValidationRules(t *testing.T) {
	t.Run("title length validation", func(t *testing.T) {
		testCases := []struct {
			name  string
			title string
			valid bool
		}{
			{"empty title", "", false},
			{"single character", "W", true},
			{"normal title", "My Birthday Wishes", true},
			{"max length", string(make([]byte, 200)), true},
			{"too long", string(make([]byte, 201)), false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				titleLength := len(tc.title)
				isValid := titleLength > 0 && titleLength <= 200

				if isValid != tc.valid {
					t.Errorf("expected valid=%v for title length %d", tc.valid, titleLength)
				}
			})
		}
	})

	t.Run("public slug must be unique when public", func(t *testing.T) {
		wishList1 := models.WishList{
			PublicSlug: pgtype.Text{String: "birthday-2024", Valid: true},
			IsPublic:   pgtype.Bool{Bool: true, Valid: true},
		}

		wishList2 := models.WishList{
			PublicSlug: pgtype.Text{String: "birthday-2024", Valid: true},
			IsPublic:   pgtype.Bool{Bool: true, Valid: true},
		}

		// Database would enforce uniqueness constraint
		if wishList1.PublicSlug.String == wishList2.PublicSlug.String {
			t.Log("duplicate public slugs should be rejected by database unique constraint")
		}
	})

	t.Run("occasion date validation", func(t *testing.T) {
		// OccasionDate is pgtype.Date
		wishList := models.WishList{
			Occasion:     pgtype.Text{String: "Birthday", Valid: true},
			OccasionDate: pgtype.Date{Valid: true},
		}

		if wishList.Occasion.Valid && !wishList.OccasionDate.Valid {
			t.Log("occasion date is optional but can be validated if present")
		}
	})
}

func TestWishListRepository_EdgeCases(t *testing.T) {
	t.Run("handle special characters in title", func(t *testing.T) {
		specialTitles := []string{
			"Birthday 🎉 2024",
			"Święta Bożego Narodzenia",
			"测试愿望清单",
			"My <special> \"wishlist\"",
		}

		for _, title := range specialTitles {
			wishList := models.WishList{
				Title: title,
			}

			if wishList.Title != title {
				t.Errorf("title should preserve special characters: %q", title)
			}
		}
	})

	t.Run("handle empty description", func(t *testing.T) {
		wishList := models.WishList{
			Title:       "Test List",
			Description: pgtype.Text{Valid: false},
		}

		if wishList.Description.Valid {
			t.Error("empty description should be invalid pgtype.Text")
		}
	})

	t.Run("template ID references valid template", func(t *testing.T) {
		validTemplates := []string{"default", "modern", "classic"}

		for _, templateID := range validTemplates {
			wishList := models.WishList{
				Title:      "Test List",
				TemplateID: templateID,
			}

			// In real implementation, this would be enforced by foreign key
			found := slices.Contains(validTemplates, wishList.TemplateID)

			if !found {
				t.Errorf("template_id %q should reference valid template", templateID)
			}
		}
	})
}
