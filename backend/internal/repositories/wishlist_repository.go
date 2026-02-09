package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	db "wish-list/internal/shared/db/models"
)

// Sentinel errors for wishlist repository
var (
	ErrWishListNotFound = errors.New("wishlist not found")
)

//go:generate go run github.com/matryer/moq@latest -out ../services/mock_wishlist_repository_test.go -pkg services . WishListRepositoryInterface

// WishListRepositoryInterface defines the interface for wishlist database operations
type WishListRepositoryInterface interface {
	Create(ctx context.Context, wishList db.WishList) (*db.WishList, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*db.WishList, error)
	GetByOwner(ctx context.Context, ownerID pgtype.UUID) ([]*db.WishList, error)
	GetByPublicSlug(ctx context.Context, publicSlug string) (*db.WishList, error)
	GetByOwnerWithItemCount(ctx context.Context, ownerID pgtype.UUID) ([]*db.WishListWithItemCount, error)
	Update(ctx context.Context, wishList db.WishList) (*db.WishList, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	DeleteWithExecutor(ctx context.Context, executor db.Executor, id pgtype.UUID) error
	IncrementViewCount(ctx context.Context, id pgtype.UUID) error
}

type WishListRepository struct {
	db *db.DB
}

func NewWishListRepository(database *db.DB) WishListRepositoryInterface {
	return &WishListRepository{
		db: database,
	}
}

// Create inserts a new wishlist into the database
func (r *WishListRepository) Create(ctx context.Context, wishList db.WishList) (*db.WishList, error) {
	query := `
		INSERT INTO wishlists (
			owner_id, title, description, occasion, occasion_date, template_id, is_public, public_slug
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING
			id, owner_id, title, description, occasion, occasion_date, template_id, is_public, public_slug, view_count, created_at, updated_at
	`

	var createdWishList db.WishList
	err := r.db.QueryRowxContext(ctx, query,
		wishList.OwnerID,
		wishList.Title,
		db.TextToString(wishList.Description),
		db.TextToString(wishList.Occasion),
		wishList.OccasionDate,
		wishList.TemplateID,
		wishList.IsPublic,
		wishList.PublicSlug, // Pass pgtype.Text directly to preserve NULL
	).StructScan(&createdWishList)

	if err != nil {
		return nil, fmt.Errorf("failed to create wishlist: %w", err)
	}

	return &createdWishList, nil
}

// GetByID retrieves a wishlist by ID
func (r *WishListRepository) GetByID(ctx context.Context, id pgtype.UUID) (*db.WishList, error) {
	query := `
		SELECT
			id, owner_id, title, description, occasion, occasion_date, template_id, is_public, public_slug, view_count, created_at, updated_at
		FROM wishlists
		WHERE id = $1
	`

	var wishList db.WishList
	err := r.db.GetContext(ctx, &wishList, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWishListNotFound
		}
		return nil, fmt.Errorf("failed to get wishlist: %w", err)
	}

	return &wishList, nil
}

// GetByPublicSlug retrieves a public wishlist by its slug
func (r *WishListRepository) GetByPublicSlug(ctx context.Context, publicSlug string) (*db.WishList, error) {
	query := `
		SELECT
			id, owner_id, title, description, occasion, occasion_date, template_id, is_public, public_slug, view_count, created_at, updated_at
		FROM wishlists
		WHERE public_slug = $1 AND is_public = true
	`

	var wishList db.WishList
	err := r.db.GetContext(ctx, &wishList, query, publicSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWishListNotFound
		}
		return nil, fmt.Errorf("failed to get wishlist by public slug: %w", err)
	}

	return &wishList, nil
}

// GetByOwner retrieves wishlists by owner ID
func (r *WishListRepository) GetByOwner(ctx context.Context, ownerID pgtype.UUID) ([]*db.WishList, error) {
	query := `
		SELECT
			id, owner_id, title, description, occasion, occasion_date, template_id, is_public, public_slug, view_count, created_at, updated_at
		FROM wishlists
		WHERE owner_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`

	var wishLists []*db.WishList
	err := r.db.SelectContext(ctx, &wishLists, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wishlists by owner: %w", err)
	}

	return wishLists, nil
}

// Update modifies an existing wishlist
func (r *WishListRepository) Update(ctx context.Context, wishList db.WishList) (*db.WishList, error) {
	query := `
		UPDATE wishlists SET
			title = $2,
			description = $3,
			occasion = $4,
			occasion_date = $5,
			template_id = $6,
			is_public = $7,
			public_slug = $8,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id, owner_id, title, description, occasion, occasion_date, template_id, is_public, public_slug, view_count, created_at, updated_at
	`

	var updatedWishList db.WishList
	err := r.db.QueryRowxContext(ctx, query,
		wishList.ID,
		wishList.Title,
		db.TextToString(wishList.Description),
		db.TextToString(wishList.Occasion),
		wishList.OccasionDate,
		wishList.TemplateID,
		wishList.IsPublic,
		wishList.PublicSlug, // Pass pgtype.Text directly to preserve NULL
	).StructScan(&updatedWishList)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWishListNotFound
		}
		return nil, fmt.Errorf("failed to update wishlist: %w", err)
	}

	return &updatedWishList, nil
}

// Delete removes a wishlist by ID
func (r *WishListRepository) Delete(ctx context.Context, id pgtype.UUID) error {
	return r.DeleteWithExecutor(ctx, r.db, id)
}

// DeleteWithExecutor removes a wishlist by ID using the provided executor (for transactions)
func (r *WishListRepository) DeleteWithExecutor(ctx context.Context, executor db.Executor, id pgtype.UUID) error {
	query := `DELETE FROM wishlists WHERE id = $1`

	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete wishlist: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrWishListNotFound
	}

	return nil
}

// IncrementViewCount increases the view count for a wishlist
func (r *WishListRepository) IncrementViewCount(ctx context.Context, id pgtype.UUID) error {
	query := `UPDATE wishlists SET view_count = view_count + 1 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrWishListNotFound
	}

	return nil
}

// GetByOwnerWithItemCount retrieves wishlists by owner ID with item counts in a single query
func (r *WishListRepository) GetByOwnerWithItemCount(ctx context.Context, ownerID pgtype.UUID) ([]*db.WishListWithItemCount, error) {
	query := `
		SELECT
			w.id, w.owner_id, w.title, w.description, w.occasion, w.occasion_date, w.template_id, w.is_public, w.public_slug, w.view_count, w.created_at, w.updated_at,
			COUNT(gi.id) AS item_count
		FROM wishlists w
		LEFT JOIN wishlist_items wi ON wi.wishlist_id = w.id
		LEFT JOIN gift_items gi ON gi.id = wi.gift_item_id AND gi.archived_at IS NULL
		WHERE w.owner_id = $1
		GROUP BY w.id, w.owner_id, w.title, w.description, w.occasion, w.occasion_date, w.template_id, w.is_public, w.public_slug, w.view_count, w.created_at, w.updated_at
		ORDER BY w.created_at DESC
		LIMIT 100
	`

	var wishLists []*db.WishListWithItemCount
	err := r.db.SelectContext(ctx, &wishLists, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wishlists by owner with item count: %w", err)
	}

	return wishLists, nil
}
