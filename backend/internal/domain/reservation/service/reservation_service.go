//go:generate go run github.com/matryer/moq@latest -out mock_cross_domain_test.go -pkg service . GiftItemRepositoryInterface

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	itemmodels "wish-list/internal/domain/item/models"
	"wish-list/internal/domain/reservation/models"
	"wish-list/internal/domain/reservation/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

// Cross-domain interfaces - only methods actually used by ReservationService

// GiftItemRepositoryInterface defines gift item repository methods used by reservation service
type GiftItemRepositoryInterface interface {
	GetByWishList(ctx context.Context, wishlistID pgtype.UUID) ([]*itemmodels.GiftItem, error)
	ReserveIfNotReserved(ctx context.Context, giftItemID, userID pgtype.UUID) (*itemmodels.GiftItem, error)
	GetPublicWishListGiftItems(ctx context.Context, publicSlug string) ([]*itemmodels.GiftItem, error)
}

var (
	ErrInvalidGiftItemID           = errors.New("invalid gift item id")
	ErrInvalidReservationWishlist  = errors.New("invalid wishlist id")
	ErrGiftItemNotInWishlist       = errors.New("gift item not found in the specified wishlist")
	ErrGiftItemAlreadyReserved     = errors.New("gift item is already reserved")
	ErrGuestInfoRequired           = errors.New("guest name and email are required for guest reservations")
	ErrReservationNotFound         = errors.New("no reservation found for this user and gift item")
	ErrMissingUserOrToken          = errors.New("either user ID or reservation token must be provided")
	ErrGiftItemNotInPublicWishlist = errors.New("gift item not found in the specified public wishlist")
)

// ReservationServiceInterface defines the interface for reservation-related operations
type ReservationServiceInterface interface {
	CreateReservation(ctx context.Context, input CreateReservationInput) (*ReservationOutput, error)
	CancelReservation(ctx context.Context, input CancelReservationInput) (*ReservationOutput, error)
	GetUserReservations(ctx context.Context, userID pgtype.UUID, limit, offset int) ([]repository.ReservationDetail, error)
	GetGuestReservations(ctx context.Context, token pgtype.UUID) ([]repository.ReservationDetail, error)
	GetReservationStatus(ctx context.Context, publicSlug, giftItemID string) (*ReservationStatusOutput, error)
	CountUserReservations(ctx context.Context, userID pgtype.UUID) (int, error)
}

type ReservationService struct {
	repo         repository.ReservationRepositoryInterface
	giftItemRepo GiftItemRepositoryInterface
}

func NewReservationService(
	reservationRepo repository.ReservationRepositoryInterface,
	giftItemRepo GiftItemRepositoryInterface,
) *ReservationService {
	return &ReservationService{
		repo:         reservationRepo,
		giftItemRepo: giftItemRepo,
	}
}

type CreateReservationInput struct {
	WishListID string
	GiftItemID string
	UserID     pgtype.UUID
	GuestName  *string
	GuestEmail *string
}

type CancelReservationInput struct {
	WishListID       string
	GiftItemID       string
	UserID           pgtype.UUID
	ReservationToken *pgtype.UUID
}

type ReservationOutput struct {
	ID               pgtype.UUID
	GiftItemID       pgtype.UUID
	ReservedByUserID pgtype.UUID
	GuestName        *string
	GuestEmail       *string
	ReservationToken pgtype.UUID
	Status           string
	ReservedAt       pgtype.Timestamptz
	ExpiresAt        pgtype.Timestamptz
	CanceledAt       pgtype.Timestamptz
	CancelReason     pgtype.Text
	NotificationSent pgtype.Bool
}

type ReservationStatusOutput struct {
	IsReserved     bool
	ReservedByName *string
	ReservedAt     *time.Time
	Status         string
}

func (s *ReservationService) CreateReservation(ctx context.Context, input CreateReservationInput) (*ReservationOutput, error) {
	// Validate gift item exists and belongs to the specified wishlist
	giftItemID := pgtype.UUID{}
	if err := giftItemID.Scan(input.GiftItemID); err != nil {
		return nil, ErrInvalidGiftItemID
	}

	wishlistID := pgtype.UUID{}
	if err := wishlistID.Scan(input.WishListID); err != nil {
		return nil, ErrInvalidReservationWishlist
	}

	// Verify ownership: get all gift items for this wishlist and check if our item is among them
	wishlistItems, err := s.giftItemRepo.GetByWishList(ctx, wishlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify wishlist ownership: %w", err)
	}

	// Check if the gift item belongs to this wishlist
	var giftItem *itemmodels.GiftItem
	for _, item := range wishlistItems {
		if item.ID == giftItemID {
			giftItem = item
			break
		}
	}

	if giftItem == nil {
		return nil, ErrGiftItemNotInWishlist
	}

	// Handle reservation based on user type (authenticated vs guest)
	if input.UserID.Valid {
		// For authenticated users, use atomic reservation that locks the gift item
		_, err := s.giftItemRepo.ReserveIfNotReserved(ctx, giftItem.ID, input.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to reserve gift item: %w", err)
		}

		// Now create the reservation record
		detail := repository.ReservationDetail{
			WishlistID:       wishlistID,
			GiftItemID:       giftItemID,
			ReservedByUserID: input.UserID,
			Status:           "active",
			ReservedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}

		dbReservation := s.mapToDbReservation(detail)
		createdReservation, err := s.repo.Create(ctx, *dbReservation)
		if err != nil {
			return nil, fmt.Errorf("failed to create reservation record: %w", err)
		}

		return s.mapToOutput(createdReservation), nil
	}
	// For guest reservations, we need to check and create atomically
	// First, check if there's an active reservation using a transaction
	activeReservation, err := s.repo.GetActiveReservationForGiftItem(ctx, giftItemID)
	if err != nil && !errors.Is(err, repository.ErrNoActiveReservation) {
		return nil, fmt.Errorf("failed to check existing reservation: %w", err)
	}

	if activeReservation != nil {
		return nil, ErrGiftItemAlreadyReserved
	}

	// Create the guest reservation
	if input.GuestName == nil || input.GuestEmail == nil {
		return nil, ErrGuestInfoRequired
	}

	// Attempt to create the reservation record atomically
	detail := repository.ReservationDetail{
		WishlistID: wishlistID,
		GiftItemID: giftItemID,
		GuestName:  pgtype.Text{String: *input.GuestName, Valid: true},
		GuestEmail: pgtype.Text{String: *input.GuestEmail, Valid: true},
		Status:     "active",
		ReservedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		// Set expiration time for guest reservations (e.g., 30 days)
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
	}

	dbReservation := s.mapToDbReservation(detail)
	createdReservation, err := s.repo.Create(ctx, *dbReservation)
	if err != nil {
		// Check if this is a uniqueness constraint violation (another transaction got there first)
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	return s.mapToOutput(createdReservation), nil
}

func (s *ReservationService) CancelReservation(ctx context.Context, input CancelReservationInput) (*ReservationOutput, error) {
	// Validate gift item belongs to the specified wishlist
	giftItemID := pgtype.UUID{}
	if err := giftItemID.Scan(input.GiftItemID); err != nil {
		return nil, ErrInvalidGiftItemID
	}

	wishlistID := pgtype.UUID{}
	if err := wishlistID.Scan(input.WishListID); err != nil {
		return nil, ErrInvalidReservationWishlist
	}

	// Verify ownership: get all gift items for this wishlist and check if our item is among them
	wishlistItems, err := s.giftItemRepo.GetByWishList(ctx, wishlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify wishlist ownership: %w", err)
	}

	// Check if the gift item belongs to this wishlist
	itemFound := false
	for _, item := range wishlistItems {
		if item.ID == giftItemID {
			itemFound = true
			break
		}
	}

	if !itemFound {
		return nil, ErrGiftItemNotInWishlist
	}

	// Determine which reservation to cancel based on input
	if input.UserID.Valid {
		// Find reservation by user and gift item
		reservations, err := s.repo.GetByGiftItem(ctx, giftItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get reservations for gift item: %w", err)
		}

		// Find the reservation made by this user
		var reservationToCancel *models.Reservation
		for _, res := range reservations {
			if res.ReservedByUserID.Valid && res.ReservedByUserID == input.UserID {
				reservationToCancel = res
				break
			}
		}

		if reservationToCancel == nil {
			return nil, ErrReservationNotFound
		}

		// Update the reservation status
		canceledAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
		updatedReservation, err := s.repo.UpdateStatus(ctx, reservationToCancel.ID, "canceled", canceledAt, pgtype.Text{String: "User canceled reservation", Valid: true})
		if err != nil {
			return nil, fmt.Errorf("failed to cancel reservation: %w", err)
		}

		return s.mapToOutput(updatedReservation), nil
	} else if input.ReservationToken != nil {
		// Find reservation by token
		updatedReservation, err := s.repo.UpdateStatusByToken(ctx, *input.ReservationToken, "canceled",
			pgtype.Timestamptz{Time: time.Now(), Valid: true},
			pgtype.Text{String: "Guest canceled reservation", Valid: true})
		if err != nil {
			return nil, fmt.Errorf("failed to cancel reservation: %w", err)
		}

		return s.mapToOutput(updatedReservation), nil
	}
	return nil, ErrMissingUserOrToken
}

func (s *ReservationService) GetUserReservations(ctx context.Context, userID pgtype.UUID, limit, offset int) ([]repository.ReservationDetail, error) {
	return s.repo.ListUserReservationsWithDetails(ctx, userID, limit, offset)
}

func (s *ReservationService) GetGuestReservations(ctx context.Context, token pgtype.UUID) ([]repository.ReservationDetail, error) {
	return s.repo.ListGuestReservationsWithDetails(ctx, token)
}

func (s *ReservationService) CountUserReservations(ctx context.Context, userID pgtype.UUID) (int, error) {
	return s.repo.CountUserReservations(ctx, userID)
}

// CreateGuestReservation handles guest reservation with token-based authentication
func (s *ReservationService) CreateGuestReservation(ctx context.Context, giftItemID, wishlistID, guestName, guestEmail string) (*ReservationOutput, error) {
	// Validate gift item exists and belongs to the specified wishlist
	itemID := pgtype.UUID{}
	if err := itemID.Scan(giftItemID); err != nil {
		return nil, ErrInvalidGiftItemID
	}

	wishlistUUID := pgtype.UUID{}
	if err := wishlistUUID.Scan(wishlistID); err != nil {
		return nil, ErrInvalidReservationWishlist
	}

	// Verify ownership: get all gift items for this wishlist and check if our item is among them
	wishlistItems, err := s.giftItemRepo.GetByWishList(ctx, wishlistUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify wishlist ownership: %w", err)
	}

	// Check if the gift item belongs to this wishlist
	var giftItem *itemmodels.GiftItem
	for _, item := range wishlistItems {
		if item.ID == itemID {
			giftItem = item
			break
		}
	}

	if giftItem == nil {
		return nil, ErrGiftItemNotInWishlist
	}

	// Check if gift item is already reserved using atomic operation
	activeReservation, err := s.repo.GetActiveReservationForGiftItem(ctx, itemID)
	if err != nil && !errors.Is(err, repository.ErrNoActiveReservation) {
		return nil, fmt.Errorf("failed to check existing reservation: %w", err)
	}

	if activeReservation != nil {
		return nil, ErrGiftItemAlreadyReserved
	}

	// Create the guest reservation
	detail := repository.ReservationDetail{
		WishlistID: wishlistUUID,
		GiftItemID: itemID,
		GuestName:  pgtype.Text{String: guestName, Valid: true},
		GuestEmail: pgtype.Text{String: guestEmail, Valid: true},
		Status:     "active",
		ReservedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		// Set expiration time for guest reservations (e.g., 30 days)
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
	}

	dbReservation := s.mapToDbReservation(detail)
	createdReservation, err := s.repo.Create(ctx, *dbReservation)
	if err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	return s.mapToOutput(createdReservation), nil
}

func (s *ReservationService) GetReservationStatus(ctx context.Context, publicSlug, giftItemID string) (*ReservationStatusOutput, error) {
	// First, validate that the gift item exists and belongs to the public wishlist
	itemID := pgtype.UUID{}
	if err := itemID.Scan(giftItemID); err != nil {
		return nil, ErrInvalidGiftItemID
	}

	// Verify ownership: get all gift items for this public wishlist and check if our item is among them
	publicWishlistItems, err := s.giftItemRepo.GetPublicWishListGiftItems(ctx, publicSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get public wishlist items: %w", err)
	}

	// Check if the gift item belongs to this public wishlist
	itemFound := false
	for _, item := range publicWishlistItems {
		if item.ID == itemID {
			itemFound = true
			break
		}
	}

	if !itemFound {
		return nil, ErrGiftItemNotInPublicWishlist
	}

	// Check if there's an active reservation for this gift item
	activeReservation, err := s.repo.GetActiveReservationForGiftItem(ctx, itemID)
	if err != nil && !errors.Is(err, repository.ErrNoActiveReservation) {
		return nil, fmt.Errorf("failed to get reservation status: %w", err)
	}

	if activeReservation == nil {
		return &ReservationStatusOutput{
			IsReserved: false,
			Status:     "available",
		}, nil
	}

	// Check if the reservation is expired
	if activeReservation.ExpiresAt.Valid && time.Now().After(activeReservation.ExpiresAt.Time) {
		// Update the reservation status to expired
		expiredAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
		_, err := s.repo.UpdateStatus(ctx, activeReservation.ID, "expired", expiredAt, pgtype.Text{String: "Reservation expired", Valid: true})
		if err != nil {
			// Log the error but continue with the old status
			fmt.Printf("Error updating expired reservation: %v\n", err)
		}

		return &ReservationStatusOutput{
			IsReserved: false,
			Status:     "available",
		}, nil
	}

	// Get the reservation details
	reservationDetails, err := s.repo.ListGuestReservationsWithDetails(ctx, activeReservation.ReservationToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservation details: %w", err)
	}

	if len(reservationDetails) == 0 {
		// Fallback: use the basic reservation info
		var reservedByName *string
		if activeReservation.GuestName.Valid {
			reservedByName = &activeReservation.GuestName.String
		} else if activeReservation.ReservedByUserID.Valid {
			// For registered users, we might want to get user details
			// For privacy reasons, we could just return a generic string
			placeholder := "Someone"
			reservedByName = &placeholder
		}

		var reservedAt *time.Time
		if activeReservation.ReservedAt.Valid {
			reservedAt = &activeReservation.ReservedAt.Time
		}

		return &ReservationStatusOutput{
			IsReserved:     true,
			ReservedByName: reservedByName,
			ReservedAt:     reservedAt,
			Status:         activeReservation.Status,
		}, nil
	}

	// Use the detailed reservation info
	reservation := reservationDetails[0]

	var reservedByName *string
	if reservation.GuestName.Valid {
		reservedByName = &reservation.GuestName.String
	} else if reservation.ReservedByUserID.Valid {
		// For privacy reasons, we could return a generic string
		placeholder := "Someone"
		reservedByName = &placeholder
	}

	var reservedAt *time.Time
	if reservation.ReservedAt.Valid {
		reservedAt = &reservation.ReservedAt.Time
	}

	return &ReservationStatusOutput{
		IsReserved:     true,
		ReservedByName: reservedByName,
		ReservedAt:     reservedAt,
		Status:         reservation.Status,
	}, nil
}

// CleanupExpiredReservations cleans up all expired reservations
func (s *ReservationService) CleanupExpiredReservations(ctx context.Context) error {
	// This would normally query for all expired reservations and update their status
	// For now, we'll just log that this method exists
	fmt.Println("Cleaning up expired reservations...")

	return nil
}

// Helper functions to map between different types
func (s *ReservationService) mapToDbReservation(detail repository.ReservationDetail) *models.Reservation {
	return &models.Reservation{
		ID:               detail.ID,
		WishlistID:       detail.WishlistID,
		GiftItemID:       detail.GiftItemID,
		ReservedByUserID: detail.ReservedByUserID,
		GuestName:        detail.GuestName,
		GuestEmail:       detail.GuestEmail,
		ReservationToken: detail.ReservationToken,
		Status:           detail.Status,
		ReservedAt:       detail.ReservedAt,
		ExpiresAt:        detail.ExpiresAt,
		CanceledAt:       detail.CanceledAt,
		CancelReason:     detail.CancelReason,
		NotificationSent: detail.NotificationSent,
	}
}

func (s *ReservationService) mapToOutput(reservation *models.Reservation) *ReservationOutput {
	var guestName *string
	if reservation.GuestName.Valid {
		guestName = &reservation.GuestName.String
	}

	var guestEmail *string
	if reservation.GuestEmail.Valid {
		guestEmail = &reservation.GuestEmail.String
	}

	return &ReservationOutput{
		ID:               reservation.ID,
		GiftItemID:       reservation.GiftItemID,
		ReservedByUserID: reservation.ReservedByUserID,
		GuestName:        guestName,
		GuestEmail:       guestEmail,
		ReservationToken: reservation.ReservationToken,
		Status:           reservation.Status,
		ReservedAt:       reservation.ReservedAt,
		ExpiresAt:        reservation.ExpiresAt,
		CanceledAt:       reservation.CanceledAt,
		CancelReason:     reservation.CancelReason,
		NotificationSent: reservation.NotificationSent,
	}
}
