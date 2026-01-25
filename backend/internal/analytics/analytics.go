package analytics

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Event types for tracking user engagement
const (
	EventUserRegistered       = "user_registered"
	EventUserLogin            = "user_login"
	EventWishListCreated      = "wishlist_created"
	EventWishListViewed       = "wishlist_viewed"
	EventWishListShared       = "wishlist_shared"
	EventGiftItemAdded        = "gift_item_added"
	EventGiftItemReserved     = "gift_item_reserved"
	EventGiftItemPurchased    = "gift_item_purchased"
	EventReservationCancelled = "reservation_cancelled"
	EventTemplateChanged      = "template_changed"
	EventAccountDeleted       = "account_deleted"
)

// Event represents an analytics event
type Event struct {
	EventType  string                 `json:"event_type"`
	UserID     string                 `json:"user_id,omitempty"`
	GuestID    string                 `json:"guest_id,omitempty"`
	Properties map[string]interface{} `json:"properties"`
	Timestamp  time.Time              `json:"timestamp"`
}

// AnalyticsService handles user engagement analytics
type AnalyticsService struct {
	// In production, this would integrate with services like:
	// - Google Analytics
	// - Mixpanel
	// - Segment
	// - Custom analytics backend
	enabled bool
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(enabled bool) *AnalyticsService {
	return &AnalyticsService{
		enabled: enabled,
	}
}

// Track sends an analytics event
func (s *AnalyticsService) Track(ctx context.Context, event Event) error {
	if !s.enabled {
		return nil
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// In production, this would send to analytics service
	// Note: Do not log Properties as they may contain PII
	log.Printf("[ANALYTICS] Event: %s, UserID: %s, Time: %s",
		event.EventType,
		event.UserID,
		//event.Properties,
		event.Timestamp.Format(time.RFC3339))

	return nil
}

// TrackUserRegistration tracks when a user registers
func (s *AnalyticsService) TrackUserRegistration(ctx context.Context, userID, email string) error {
	return s.Track(ctx, Event{
		EventType: EventUserRegistered,
		UserID:    userID,
		Properties: map[string]interface{}{
			"email": email,
		},
	})
}

// TrackUserLogin tracks when a user logs in
func (s *AnalyticsService) TrackUserLogin(ctx context.Context, userID string) error {
	return s.Track(ctx, Event{
		EventType: EventUserLogin,
		UserID:    userID,
		Properties: map[string]interface{}{
			"login_method": "email_password",
		},
	})
}

// TrackWishListCreated tracks when a wishlist is created
func (s *AnalyticsService) TrackWishListCreated(ctx context.Context, userID, wishListID, templateID string, isPublic bool) error {
	return s.Track(ctx, Event{
		EventType: EventWishListCreated,
		UserID:    userID,
		Properties: map[string]interface{}{
			"wishlist_id": wishListID,
			"template_id": templateID,
			"is_public":   isPublic,
		},
	})
}

// TrackWishListViewed tracks when a wishlist is viewed
func (s *AnalyticsService) TrackWishListViewed(ctx context.Context, wishListID string, userID string, isOwner bool) error {
	return s.Track(ctx, Event{
		EventType: EventWishListViewed,
		UserID:    userID,
		Properties: map[string]interface{}{
			"wishlist_id": wishListID,
			"is_owner":    isOwner,
		},
	})
}

// TrackWishListShared tracks when a wishlist is shared
func (s *AnalyticsService) TrackWishListShared(ctx context.Context, userID, wishListID, shareMethod string) error {
	return s.Track(ctx, Event{
		EventType: EventWishListShared,
		UserID:    userID,
		Properties: map[string]interface{}{
			"wishlist_id":  wishListID,
			"share_method": shareMethod, // "link", "email", etc.
		},
	})
}

// TrackGiftItemAdded tracks when a gift item is added
func (s *AnalyticsService) TrackGiftItemAdded(ctx context.Context, userID, wishListID, giftItemID string, hasImage bool) error {
	return s.Track(ctx, Event{
		EventType: EventGiftItemAdded,
		UserID:    userID,
		Properties: map[string]interface{}{
			"wishlist_id":  wishListID,
			"gift_item_id": giftItemID,
			"has_image":    hasImage,
		},
	})
}

// TrackGiftItemReserved tracks when a gift item is reserved
func (s *AnalyticsService) TrackGiftItemReserved(ctx context.Context, userID, guestID, giftItemID string, isGuest bool) error {
	event := Event{
		EventType: EventGiftItemReserved,
		Properties: map[string]interface{}{
			"gift_item_id": giftItemID,
			"is_guest":     isGuest,
		},
	}

	if isGuest {
		event.GuestID = guestID
	} else {
		event.UserID = userID
	}

	return s.Track(ctx, event)
}

// TrackGiftItemPurchased tracks when a gift item is marked as purchased
func (s *AnalyticsService) TrackGiftItemPurchased(ctx context.Context, userID, giftItemID string, price float64) error {
	return s.Track(ctx, Event{
		EventType: EventGiftItemPurchased,
		UserID:    userID,
		Properties: map[string]interface{}{
			"gift_item_id":    giftItemID,
			"purchased_price": price,
		},
	})
}

// TrackReservationCancelled tracks when a reservation is cancelled
func (s *AnalyticsService) TrackReservationCancelled(ctx context.Context, userID, giftItemID, reason string) error {
	return s.Track(ctx, Event{
		EventType: EventReservationCancelled,
		UserID:    userID,
		Properties: map[string]interface{}{
			"gift_item_id": giftItemID,
			"reason":       reason,
		},
	})
}

// TrackTemplateChanged tracks when a wishlist template is changed
func (s *AnalyticsService) TrackTemplateChanged(ctx context.Context, userID, wishListID, oldTemplateID, newTemplateID string) error {
	return s.Track(ctx, Event{
		EventType: EventTemplateChanged,
		UserID:    userID,
		Properties: map[string]interface{}{
			"wishlist_id":     wishListID,
			"old_template_id": oldTemplateID,
			"new_template_id": newTemplateID,
		},
	})
}

// TrackAccountDeleted tracks when an account is deleted
func (s *AnalyticsService) TrackAccountDeleted(ctx context.Context, userID string, reason string, isAutomatic bool) error {
	return s.Track(ctx, Event{
		EventType: EventAccountDeleted,
		UserID:    userID,
		Properties: map[string]interface{}{
			"reason":       reason,
			"is_automatic": isAutomatic,
		},
	})
}

// GetEngagementMetrics would return aggregated engagement metrics
// In production, this would query the analytics backend
func (s *AnalyticsService) GetEngagementMetrics(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	if !s.enabled {
		return map[string]interface{}{
			"message": "Analytics disabled",
		}, nil
	}

	// In production, this would return real metrics
	return map[string]interface{}{
		"total_users":           0,
		"active_users":          0,
		"wishlists_created":     0,
		"gifts_reserved":        0,
		"gifts_purchased":       0,
		"public_wishlist_views": 0,
		"period_start":          startDate.Format(time.RFC3339),
		"period_end":            endDate.Format(time.RFC3339),
		"note":                  "Production implementation would query analytics backend",
	}, nil
}

// BatchTrack sends multiple events in a batch
func (s *AnalyticsService) BatchTrack(ctx context.Context, events []Event) error {
	if !s.enabled {
		return nil
	}

	for _, event := range events {
		if err := s.Track(ctx, event); err != nil {
			// Log error but continue processing other events
			fmt.Printf("Error tracking event %s: %v\n", event.EventType, err)
		}
	}

	return nil
}
