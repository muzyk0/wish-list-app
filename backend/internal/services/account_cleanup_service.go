package services

import (
	"context"
	"fmt"
	"log"
	"time"
	db "wish-list/internal/db/models"
	"wish-list/internal/repositories"

	"github.com/jackc/pgx/v5/pgtype"
)

// AccountCleanupService handles account inactivity tracking and deletion
type AccountCleanupService struct {
	userRepo        repositories.UserRepositoryInterface
	wishListRepo    repositories.WishListRepositoryInterface
	giftItemRepo    repositories.GiftItemRepositoryInterface
	reservationRepo repositories.ReservationRepositoryInterface
	emailService    EmailServiceInterface
}

// NewAccountCleanupService creates a new account cleanup service
func NewAccountCleanupService(
	userRepo repositories.UserRepositoryInterface,
	wishListRepo repositories.WishListRepositoryInterface,
	giftItemRepo repositories.GiftItemRepositoryInterface,
	reservationRepo repositories.ReservationRepositoryInterface,
	emailService EmailServiceInterface,
) *AccountCleanupService {
	return &AccountCleanupService{
		userRepo:        userRepo,
		wishListRepo:    wishListRepo,
		giftItemRepo:    giftItemRepo,
		reservationRepo: reservationRepo,
		emailService:    emailService,
	}
}

// CheckInactiveAccounts identifies accounts approaching the deletion threshold
// and sends warning notifications
func (s *AccountCleanupService) CheckInactiveAccounts(ctx context.Context) error {
	now := time.Now()

	// Check for accounts inactive for 23 months (1 month before deletion)
	threshold23Months := now.AddDate(0, -23, 0)
	inactiveUsers23, err := s.findInactiveUsersSince(ctx, threshold23Months)
	if err != nil {
		return fmt.Errorf("failed to find 23-month inactive users: %w", err)
	}

	for _, user := range inactiveUsers23 {
		// Send 1-month warning
		userName := user.FirstName.String
		if user.LastName.Valid {
			userName += " " + user.LastName.String
		}
		if err := s.emailService.SendAccountInactivityNotification(ctx, user.Email, userName); err != nil {
			log.Printf("Failed to send 23-month warning to user %s: %v", user.ID.String(), err)
		}
		log.Printf("Sent 23-month inactivity warning to user %s", user.ID.String())
	}

	// Check for accounts 7 days before 24 months
	threshold24MonthsMinus7Days := now.AddDate(0, -24, 0).Add(7 * 24 * time.Hour)
	inactiveUsers7Days, err := s.findInactiveUsersSince(ctx, threshold24MonthsMinus7Days)
	if err != nil {
		return fmt.Errorf("failed to find users 7 days from deletion: %w", err)
	}

	for _, user := range inactiveUsers7Days {
		// Send final 7-day warning
		userName := user.FirstName.String
		if user.LastName.Valid {
			userName += " " + user.LastName.String
		}
		if err := s.emailService.SendAccountInactivityNotification(ctx, user.Email, userName); err != nil {
			log.Printf("Failed to send 7-day warning to user %s: %v", user.ID.String(), err)
		}
		log.Printf("Sent 7-day inactivity warning to user %s", user.ID.String())
	}

	return nil
}

// DeleteInactiveAccounts deletes accounts that have been inactive for 24 months
func (s *AccountCleanupService) DeleteInactiveAccounts(ctx context.Context) error {
	now := time.Now()
	threshold24Months := now.AddDate(0, -24, 0)

	inactiveUsers, err := s.findInactiveUsersSince(ctx, threshold24Months)
	if err != nil {
		return fmt.Errorf("failed to find inactive users for deletion: %w", err)
	}

	for _, user := range inactiveUsers {
		log.Printf("Deleting inactive user account: %s (last active: %s)", user.ID.String(), user.UpdatedAt.Time.Format(time.RFC3339))

		if err := s.DeleteUserAccount(ctx, user.ID.String(), "automatic_inactivity_deletion"); err != nil {
			log.Printf("Failed to delete user %s: %v", user.ID.String(), err)
			continue
		}

		log.Printf("Successfully deleted inactive user %s", user.ID.String())
	}

	return nil
}

// DeleteUserAccount deletes a user account and all associated data
func (s *AccountCleanupService) DeleteUserAccount(ctx context.Context, userID string, reason string) error {
	id := pgtype.UUID{}
	if err := id.Scan(userID); err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	// Get user details for logging
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Get all user's wishlists
	wishLists, err := s.wishListRepo.GetByOwner(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user wishlists: %w", err)
	}

	// For each wishlist, notify reservation holders before deletion
	for _, wishList := range wishLists {
		// Get all gift items in this wishlist
		giftItems, err := s.giftItemRepo.GetByWishList(ctx, wishList.ID)
		if err != nil {
			log.Printf("Warning: failed to get gift items for wishlist %s: %v", wishList.ID.String(), err)
			continue
		}

		// For each gift item, check for active reservations and notify
		for _, giftItem := range giftItems {
			reservation, err := s.reservationRepo.GetActiveReservationForGiftItem(ctx, giftItem.ID)
			if err == nil && reservation != nil {
				// Notify reservation holder
				var recipientEmail string
				if reservation.GuestEmail.Valid {
					recipientEmail = reservation.GuestEmail.String
				}

				if recipientEmail != "" {
					err := s.emailService.SendReservationRemovedEmail(ctx, recipientEmail, giftItem.Name, wishList.Title)
					if err != nil {
						log.Printf("Warning: failed to send deletion notification: %v", err)
					}
				}
			}

			// Delete gift item (cascade will handle reservations)
			if err := s.giftItemRepo.Delete(ctx, giftItem.ID); err != nil {
				log.Printf("Warning: failed to delete gift item %s: %v", giftItem.ID.String(), err)
			}
		}

		// Delete wishlist
		if err := s.wishListRepo.Delete(ctx, wishList.ID); err != nil {
			log.Printf("Warning: failed to delete wishlist %s: %v", wishList.ID.String(), err)
		}
	}

	// Delete user account
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Log the deletion for audit purposes
	s.logAccountDeletion(user.ID.String(), user.Email, reason, reason == "automatic_inactivity_deletion")

	return nil
}

// ExportUserData exports all user data for GDPR compliance
func (s *AccountCleanupService) ExportUserData(ctx context.Context, userID string) (map[string]interface{}, error) {
	id := pgtype.UUID{}
	if err := id.Scan(userID); err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	// Get user details
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get all wishlists
	wishLists, err := s.wishListRepo.GetByOwner(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get wishlists: %w", err)
	}

	wishListsData := make([]map[string]interface{}, 0)
	for _, wl := range wishLists {
		giftItems, _ := s.giftItemRepo.GetByWishList(ctx, wl.ID)

		giftItemsData := make([]map[string]interface{}, 0)
		for _, item := range giftItems {
			giftItemsData = append(giftItemsData, map[string]interface{}{
				"id":          item.ID.String(),
				"name":        item.Name,
				"description": item.Description.String,
				"link":        item.Link.String,
				"image_url":   item.ImageUrl.String,
				"price":       db.NumericToFloat64(item.Price),
				"priority":    item.Priority.Int32,
				"created_at":  item.CreatedAt.Time.Format(time.RFC3339),
			})
		}

		wishListsData = append(wishListsData, map[string]interface{}{
			"id":          wl.ID.String(),
			"title":       wl.Title,
			"description": wl.Description.String,
			"occasion":    wl.Occasion.String,
			"is_public":   wl.IsPublic.Bool,
			"public_slug": wl.PublicSlug.String,
			"created_at":  wl.CreatedAt.Time.Format(time.RFC3339),
			"gift_items":  giftItemsData,
		})
	}

	userName := user.FirstName.String
	if user.LastName.Valid {
		userName += " " + user.LastName.String
	}

	return map[string]interface{}{
		"user": map[string]interface{}{
			"id":         user.ID.String(),
			"email":      user.Email,
			"name":       userName,
			"created_at": user.CreatedAt.Time.Format(time.RFC3339),
			"updated_at": user.UpdatedAt.Time.Format(time.RFC3339),
		},
		"wishlists":     wishListsData,
		"exported_at":   time.Now().Format(time.RFC3339),
		"export_format": "json",
	}, nil
}

// findInactiveUsersSince finds users who haven't been active since the given date
func (s *AccountCleanupService) findInactiveUsersSince(ctx context.Context, since time.Time) ([]*db.User, error) {
	// In production, this would query users where last_login_at < since
	// For now, we return empty list as we don't have last_login_at field yet
	// This is a placeholder for the actual implementation
	return []*db.User{}, nil
}

// logAccountDeletion logs account deletion for audit purposes
func (s *AccountCleanupService) logAccountDeletion(userID, email, reason string, isAutomatic bool) {
	log.Printf("[AUDIT] Account deleted: UserID=%s, Email=%s, Reason=%s, Automatic=%v, Timestamp=%s",
		userID,
		email,
		reason,
		isAutomatic,
		time.Now().Format(time.RFC3339))
}

// StartScheduledCleanup starts the scheduled cleanup job
func (s *AccountCleanupService) StartScheduledCleanup() {
	// Run cleanup daily at 2 AM
	ticker := time.NewTicker(24 * time.Hour)

	go func() {
		for range ticker.C {
			ctx := context.Background()

			log.Println("Running scheduled account cleanup check...")

			// Check for inactive accounts and send warnings
			if err := s.CheckInactiveAccounts(ctx); err != nil {
				log.Printf("Error checking inactive accounts: %v", err)
			}

			// Delete accounts inactive for 24 months
			if err := s.DeleteInactiveAccounts(ctx); err != nil {
				log.Printf("Error deleting inactive accounts: %v", err)
			}

			log.Println("Scheduled account cleanup completed")
		}
	}()

	log.Println("Scheduled account cleanup job started (runs daily at current time)")
}
