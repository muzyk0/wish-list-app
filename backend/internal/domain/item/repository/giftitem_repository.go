//go:generate go run github.com/matryer/moq@latest -out ../service/mock_giftitem_repository_test.go -pkg service . GiftItemRepositoryInterface

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"wish-list/internal/app/database"
	"wish-list/internal/domain/item/models"
	reservationmodels "wish-list/internal/domain/reservation/models"
)

// Sentinel errors for gift item repository
var (
	ErrGiftItemNotFound          = errors.New("gift item not found")
	ErrGiftItemAlreadyReserved   = errors.New("gift item is already reserved")
	ErrGiftItemAlreadyArchived   = errors.New("item not found or already archived")
	ErrGiftItemConcurrentReserve = errors.New("gift item was reserved by another transaction")
	ErrInvalidSortField          = errors.New("invalid sort field")
	ErrInvalidSortOrder          = errors.New("invalid sort order")
)

// validSortFields defines allowed sort fields for SQL queries
var validSortFields = map[string]string{
	"created_at": "created_at",
	"updated_at": "updated_at",
	"title":      "name",
	"price":      "price",
}

// validSortOrders defines allowed sort orders for SQL queries
var validSortOrders = map[string]bool{
	"ASC":  true,
	"DESC": true,
}

// giftItemColumns is the standard column list for gift_items queries
const giftItemColumns = `id, owner_id, name, description, link, image_url, price, priority,
	reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at,
	purchased_price, notes, position, archived_at, created_at, updated_at`

// giftItemColumnsAliased is the column list prefixed with gi. alias
const giftItemColumnsAliased = `gi.id, gi.owner_id, gi.name, gi.description, gi.link, gi.image_url,
	gi.price, gi.priority, gi.reserved_by_user_id, gi.reserved_at,
	gi.purchased_by_user_id, gi.purchased_at, gi.purchased_price,
	gi.notes, gi.position, gi.archived_at, gi.created_at, gi.updated_at`

// ItemFilters contains filter and pagination parameters for querying items
type ItemFilters struct {
	Page            int
	Limit           int
	Sort            string // created_at, updated_at, title, price
	Order           string // asc, desc
	Unattached      bool   // Items not attached to any wishlist
	IncludeArchived bool   // Include archived items
	Search          string // Search in title and description
}

// PaginatedResult represents paginated query result
type PaginatedResult struct {
	Items      []*models.GiftItem
	TotalCount int64
}

// GiftItemRepositoryInterface defines the interface for gift item database operations
type GiftItemRepositoryInterface interface {
	// CRUD
	CreateWithOwner(ctx context.Context, giftItem models.GiftItem) (*models.GiftItem, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error)
	GetByOwnerPaginated(ctx context.Context, ownerID pgtype.UUID, filters ItemFilters) (*PaginatedResult, error)
	GetByWishList(ctx context.Context, wishlistID pgtype.UUID) ([]*models.GiftItem, error)
	GetPublicWishListGiftItems(ctx context.Context, publicSlug string) ([]*models.GiftItem, error)
	GetPublicWishListGiftItemsPaginated(ctx context.Context, publicSlug string, limit, offset int) ([]*models.GiftItem, int, error)
	GetUnattached(ctx context.Context, ownerID pgtype.UUID) ([]*models.GiftItem, error)
	Update(ctx context.Context, giftItem models.GiftItem) (*models.GiftItem, error)
	UpdateWithNewSchema(ctx context.Context, giftItem *models.GiftItem) (*models.GiftItem, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	DeleteWithExecutor(ctx context.Context, executor database.Executor, id pgtype.UUID) error
	SoftDelete(ctx context.Context, id pgtype.UUID) error

	// Reservation operations
	Reserve(ctx context.Context, giftItemID, userID pgtype.UUID) (*models.GiftItem, error)
	Unreserve(ctx context.Context, giftItemID pgtype.UUID) (*models.GiftItem, error)
	MarkAsPurchased(ctx context.Context, giftItemID, userID pgtype.UUID, purchasedPrice pgtype.Numeric) (*models.GiftItem, error)
	ReserveIfNotReserved(ctx context.Context, giftItemID, userID pgtype.UUID) (*models.GiftItem, error)
	DeleteWithReservationNotification(ctx context.Context, giftItemID pgtype.UUID) ([]*reservationmodels.Reservation, error)
}

// GiftItemRepository implements GiftItemRepositoryInterface
type GiftItemRepository struct {
	db *database.DB
}

// NewGiftItemRepository creates a new GiftItemRepository
func NewGiftItemRepository(db *database.DB) GiftItemRepositoryInterface {
	return &GiftItemRepository{
		db: db,
	}
}

// ---------------------------------------------------------------------------
// CRUD operations
// ---------------------------------------------------------------------------

// CreateWithOwner creates a new item with owner_id
func (r *GiftItemRepository) CreateWithOwner(ctx context.Context, giftItem models.GiftItem) (*models.GiftItem, error) {
	query := fmt.Sprintf(`
		INSERT INTO gift_items (
			owner_id, name, description, link, image_url, price, priority, notes, position
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING %s
	`, giftItemColumns)

	var created models.GiftItem
	err := r.db.GetContext(
		ctx,
		&created,
		query,
		giftItem.OwnerID,
		giftItem.Name,
		giftItem.Description,
		giftItem.Link,
		giftItem.ImageUrl,
		giftItem.Price,
		giftItem.Priority,
		giftItem.Notes,
		giftItem.Position,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gift item: %w", err)
	}

	return &created, nil
}

// GetByID retrieves a gift item by ID
func (r *GiftItemRepository) GetByID(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM gift_items
		WHERE id = $1
	`, giftItemColumns)

	var giftItem models.GiftItem
	err := r.db.GetContext(ctx, &giftItem, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGiftItemNotFound
		}
		return nil, fmt.Errorf("failed to get gift item: %w", err)
	}

	return &giftItem, nil
}

// GetByOwnerPaginated retrieves items owned by user with pagination and filters
func (r *GiftItemRepository) GetByOwnerPaginated(ctx context.Context, ownerID pgtype.UUID, filters ItemFilters) (*PaginatedResult, error) {
	whereConditions := []string{"owner_id = $1"}
	args := []any{ownerID}
	argIndex := 2

	if !filters.IncludeArchived {
		whereConditions = append(whereConditions, "archived_at IS NULL")
	}

	if filters.Unattached {
		whereConditions = append(whereConditions, `
			NOT EXISTS (
				SELECT 1 FROM wishlist_items wi
				WHERE wi.gift_item_id = gift_items.id
			)
		`)
	}

	if filters.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filters.Search+"%")
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Validate sort field against whitelist
	sortField, ok := validSortFields[filters.Sort]
	if !ok {
		return nil, ErrInvalidSortField
	}

	// Normalize and validate sort order
	order := strings.ToUpper(filters.Order)
	if order == "" {
		order = "DESC"
	}
	if !validSortOrders[order] {
		return nil, ErrInvalidSortOrder
	}

	orderClause := fmt.Sprintf("%s %s", sortField, order)
	offset := (filters.Page - 1) * filters.Limit

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM gift_items WHERE %s`, whereClause)

	var totalCount int64
	if err := r.db.GetContext(ctx, &totalCount, countQuery, args...); err != nil {
		return nil, fmt.Errorf("failed to count items: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM gift_items
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, giftItemColumns, whereClause, orderClause, argIndex, argIndex+1)

	args = append(args, filters.Limit, offset)

	var items []*models.GiftItem
	if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	return &PaginatedResult{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}

// GetByWishList retrieves gift items by wishlist ID via the wishlist_items junction table
func (r *GiftItemRepository) GetByWishList(ctx context.Context, wishlistID pgtype.UUID) ([]*models.GiftItem, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM gift_items gi
		INNER JOIN wishlist_items wi ON wi.gift_item_id = gi.id
		WHERE wi.wishlist_id = $1
		  AND gi.archived_at IS NULL
		ORDER BY gi.position ASC
		LIMIT 100
	`, giftItemColumnsAliased)

	var giftItems []*models.GiftItem
	err := r.db.SelectContext(ctx, &giftItems, query, wishlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get gift items by wishlist: %w", err)
	}

	return giftItems, nil
}

// GetPublicWishListGiftItems retrieves gift items for a public wishlist by slug
func (r *GiftItemRepository) GetPublicWishListGiftItems(ctx context.Context, publicSlug string) ([]*models.GiftItem, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM gift_items gi
		INNER JOIN wishlist_items wi ON wi.gift_item_id = gi.id
		INNER JOIN wishlists w ON wi.wishlist_id = w.id
		WHERE w.public_slug = $1 AND w.is_public = true
		  AND gi.archived_at IS NULL
		ORDER BY gi.position ASC
		LIMIT 100
	`, giftItemColumnsAliased)

	var giftItems []*models.GiftItem
	err := r.db.SelectContext(ctx, &giftItems, query, publicSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get public wishlist gift items: %w", err)
	}

	return giftItems, nil
}

// GetPublicWishListGiftItemsPaginated retrieves paginated gift items for a public wishlist by slug
// Returns the items, total count, and any error
func (r *GiftItemRepository) GetPublicWishListGiftItemsPaginated(ctx context.Context, publicSlug string, limit, offset int) ([]*models.GiftItem, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM gift_items gi
		INNER JOIN wishlist_items wi ON wi.gift_item_id = gi.id
		INNER JOIN wishlists w ON wi.wishlist_id = w.id
		WHERE w.public_slug = $1 AND w.is_public = true
		  AND gi.archived_at IS NULL
	`
	var totalCount int
	if err := r.db.GetContext(ctx, &totalCount, countQuery, publicSlug); err != nil {
		return nil, 0, fmt.Errorf("failed to count public wishlist gift items: %w", err)
	}

	// Get paginated items
	query := fmt.Sprintf(`
		SELECT %s
		FROM gift_items gi
		INNER JOIN wishlist_items wi ON wi.gift_item_id = gi.id
		INNER JOIN wishlists w ON wi.wishlist_id = w.id
		WHERE w.public_slug = $1 AND w.is_public = true
		  AND gi.archived_at IS NULL
		ORDER BY gi.position ASC
		LIMIT $2 OFFSET $3
	`, giftItemColumnsAliased)

	var giftItems []*models.GiftItem
	if err := r.db.SelectContext(ctx, &giftItems, query, publicSlug, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to get public wishlist gift items: %w", err)
	}

	return giftItems, totalCount, nil
}

// GetUnattached retrieves items not attached to any wishlist
func (r *GiftItemRepository) GetUnattached(ctx context.Context, ownerID pgtype.UUID) ([]*models.GiftItem, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM gift_items gi
		WHERE gi.owner_id = $1
		  AND gi.archived_at IS NULL
		  AND NOT EXISTS (
			  SELECT 1 FROM wishlist_items wi
			  WHERE wi.gift_item_id = gi.id
		  )
		ORDER BY gi.created_at DESC
	`, giftItemColumnsAliased)

	var items []*models.GiftItem
	if err := r.db.SelectContext(ctx, &items, query, ownerID); err != nil {
		return nil, fmt.Errorf("failed to get unattached items: %w", err)
	}

	return items, nil
}

// Update modifies an existing gift item (basic fields only)
func (r *GiftItemRepository) Update(ctx context.Context, giftItem models.GiftItem) (*models.GiftItem, error) {
	query := fmt.Sprintf(`
		UPDATE gift_items SET
			name = $2,
			description = $3,
			link = $4,
			image_url = $5,
			price = $6,
			priority = $7,
			notes = $8,
			position = $9,
			updated_at = NOW()
		WHERE id = $1 AND archived_at IS NULL
		RETURNING %s
	`, giftItemColumns)

	var updatedGiftItem models.GiftItem
	err := r.db.QueryRowxContext(ctx, query,
		giftItem.ID,
		giftItem.Name,
		database.TextToString(giftItem.Description),
		database.TextToString(giftItem.Link),
		database.TextToString(giftItem.ImageUrl),
		giftItem.Price,
		giftItem.Priority,
		database.TextToString(giftItem.Notes),
		giftItem.Position,
	).StructScan(&updatedGiftItem)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGiftItemNotFound
		}
		return nil, fmt.Errorf("failed to update gift item: %w", err)
	}

	return &updatedGiftItem, nil
}

// UpdateWithNewSchema updates an item including reservation/purchase fields
func (r *GiftItemRepository) UpdateWithNewSchema(ctx context.Context, giftItem *models.GiftItem) (*models.GiftItem, error) {
	query := fmt.Sprintf(`
		UPDATE gift_items
		SET
			name = $2,
			description = $3,
			link = $4,
			image_url = $5,
			price = $6,
			priority = $7,
			notes = $8,
			position = $9,
			reserved_by_user_id = $10,
			reserved_at = $11,
			purchased_by_user_id = $12,
			purchased_at = $13,
			purchased_price = $14,
			updated_at = $15
		WHERE id = $1 AND archived_at IS NULL
		RETURNING %s
	`, giftItemColumns)

	var updated models.GiftItem
	err := r.db.GetContext(
		ctx,
		&updated,
		query,
		giftItem.ID,
		giftItem.Name,
		giftItem.Description,
		giftItem.Link,
		giftItem.ImageUrl,
		giftItem.Price,
		giftItem.Priority,
		giftItem.Notes,
		giftItem.Position,
		giftItem.ReservedByUserID,
		giftItem.ReservedAt,
		giftItem.PurchasedByUserID,
		giftItem.PurchasedAt,
		giftItem.PurchasedPrice,
		time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update gift item: %w", err)
	}

	return &updated, nil
}

// Delete removes a gift item by ID
func (r *GiftItemRepository) Delete(ctx context.Context, id pgtype.UUID) error {
	return r.DeleteWithExecutor(ctx, r.db, id)
}

// DeleteWithExecutor removes a gift item by ID using the provided executor (for transactions)
func (r *GiftItemRepository) DeleteWithExecutor(ctx context.Context, executor database.Executor, id pgtype.UUID) error {
	query := `DELETE FROM gift_items WHERE id = $1`

	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete gift item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrGiftItemNotFound
	}

	return nil
}

// SoftDelete marks an item as archived by setting archived_at timestamp
func (r *GiftItemRepository) SoftDelete(ctx context.Context, id pgtype.UUID) error {
	query := `
		UPDATE gift_items
		SET archived_at = $1, updated_at = $2
		WHERE id = $3 AND archived_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to archive item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrGiftItemAlreadyArchived
	}

	return nil
}

// ---------------------------------------------------------------------------
// Reservation operations
// ---------------------------------------------------------------------------

// Reserve marks a gift item as reserved by a user
func (r *GiftItemRepository) Reserve(ctx context.Context, giftItemID, userID pgtype.UUID) (*models.GiftItem, error) {
	query := fmt.Sprintf(`
		UPDATE gift_items SET
			reserved_by_user_id = $2,
			reserved_at = $3,
			updated_at = NOW()
		WHERE id = $1
		RETURNING %s
	`, giftItemColumns)

	var updatedGiftItem models.GiftItem
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	err := r.db.QueryRowxContext(ctx, query,
		giftItemID,
		userID,
		now,
	).StructScan(&updatedGiftItem)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGiftItemNotFound
		}
		return nil, fmt.Errorf("failed to reserve gift item: %w", err)
	}

	return &updatedGiftItem, nil
}

// Unreserve removes reservation from a gift item
func (r *GiftItemRepository) Unreserve(ctx context.Context, giftItemID pgtype.UUID) (*models.GiftItem, error) {
	query := fmt.Sprintf(`
		UPDATE gift_items SET
			reserved_by_user_id = NULL,
			reserved_at = NULL,
			updated_at = NOW()
		WHERE id = $1
		RETURNING %s
	`, giftItemColumns)

	var updatedGiftItem models.GiftItem
	err := r.db.QueryRowxContext(ctx, query, giftItemID).StructScan(&updatedGiftItem)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGiftItemNotFound
		}
		return nil, fmt.Errorf("failed to unreserve gift item: %w", err)
	}

	return &updatedGiftItem, nil
}

// MarkAsPurchased marks a gift item as purchased
func (r *GiftItemRepository) MarkAsPurchased(ctx context.Context, giftItemID, userID pgtype.UUID, purchasedPrice pgtype.Numeric) (*models.GiftItem, error) {
	query := fmt.Sprintf(`
		UPDATE gift_items SET
			purchased_by_user_id = $2,
			purchased_at = $3,
			purchased_price = $4,
			reserved_by_user_id = NULL,
			reserved_at = NULL,
			updated_at = NOW()
		WHERE id = $1
		RETURNING %s
	`, giftItemColumns)

	var updatedGiftItem models.GiftItem
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	err := r.db.QueryRowxContext(ctx, query,
		giftItemID,
		userID,
		now,
		purchasedPrice,
	).StructScan(&updatedGiftItem)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGiftItemNotFound
		}
		return nil, fmt.Errorf("failed to mark gift item as purchased: %w", err)
	}

	return &updatedGiftItem, nil
}

// ReserveIfNotReserved atomically reserves a gift item if it's not already reserved
func (r *GiftItemRepository) ReserveIfNotReserved(ctx context.Context, giftItemID, userID pgtype.UUID) (*models.GiftItem, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			log.Printf("tx rollback error: %v", rbErr)
		}
	}()

	lockQuery := `
		SELECT id, reserved_by_user_id, reserved_at
		FROM gift_items
		WHERE id = $1
		FOR UPDATE
	`

	var currentItem models.GiftItem
	err = tx.GetContext(ctx, &currentItem, lockQuery, giftItemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGiftItemNotFound
		}
		return nil, fmt.Errorf("failed to lock gift item: %w", err)
	}

	if currentItem.ReservedByUserID.Valid {
		return nil, ErrGiftItemAlreadyReserved
	}

	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	updateQuery := fmt.Sprintf(`
		UPDATE gift_items SET
			reserved_by_user_id = $2,
			reserved_at = $3,
			updated_at = NOW()
		WHERE id = $1 AND reserved_by_user_id IS NULL
		RETURNING %s
	`, giftItemColumns)

	var updatedGiftItem models.GiftItem
	err = tx.QueryRowxContext(ctx, updateQuery,
		giftItemID,
		userID,
		now,
	).StructScan(&updatedGiftItem)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGiftItemConcurrentReserve
		}
		return nil, fmt.Errorf("failed to reserve gift item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit reservation: %w", err)
	}

	return &updatedGiftItem, nil
}

// DeleteWithReservationNotification deletes a gift item and returns any active reservations for notification purposes
func (r *GiftItemRepository) DeleteWithReservationNotification(ctx context.Context, giftItemID pgtype.UUID) ([]*reservationmodels.Reservation, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			log.Printf("tx rollback error: %v", rbErr)
		}
	}()

	getReservationsQuery := `
		SELECT id, wishlist_id, gift_item_id, reserved_by_user_id, guest_name, guest_email,
			reservation_token, status, reserved_at, expires_at, canceled_at,
			cancel_reason, notification_sent, updated_at
		FROM reservations
		WHERE gift_item_id = $1 AND status = 'active'
	`

	var activeReservations []*reservationmodels.Reservation
	err = tx.SelectContext(ctx, &activeReservations, getReservationsQuery, giftItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active reservations: %w", err)
	}

	deleteQuery := `DELETE FROM gift_items WHERE id = $1`
	result, err := tx.ExecContext(ctx, deleteQuery, giftItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete gift item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, ErrGiftItemNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return activeReservations, nil
}
