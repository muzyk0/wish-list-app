package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"time"
)

// EmailServiceInterface defines the interface for email operations
type EmailServiceInterface interface {
	SendReservationCancellationEmail(ctx context.Context, recipientEmail, giftItemName, wishlistTitle string) error
	SendReservationRemovedEmail(ctx context.Context, recipientEmail, giftItemName, wishlistTitle string) error
	SendGiftPurchasedConfirmationEmail(ctx context.Context, recipientEmail, giftItemName, wishlistTitle, guestName string) error
	SendAccountInactivityNotification(ctx context.Context, recipientEmail, userName string) error
	ScheduleAccountCleanupNotifications() // Schedules periodic checks for inactive accounts
}

type EmailService struct {
	// In a real implementation, this would contain SMTP configuration, etc.
}

func NewEmailService() *EmailService {
	return &EmailService{}
}

type ReservationCancellationEmailData struct {
	GiftItemName  string
	WishlistTitle string
}

type ReservationRemovedEmailData struct {
	GiftItemName  string
	WishlistTitle string
}

type AccountInactivityNotificationData struct {
	UserName string
}

type GiftPurchasedConfirmationEmailData struct {
	GiftItemName  string
	WishlistTitle string
	GuestName     string
}

func (s *EmailService) SendAccountInactivityNotification(ctx context.Context, recipientEmail, userName string) error {
	subject := "Account inactivity notice - scheduled deletion"
	_, err := s.buildAccountInactivityNotification(userName)
	if err != nil {
		return fmt.Errorf("failed to build email body: %w", err)
	}

	// In a real implementation, this would send the email via SMTP
	// Do not log PII (email addresses) or full body content
	log.Printf("Email send simulated: subject=%q (recipient redacted)", subject)

	return nil
}

func (s *EmailService) ScheduleAccountCleanupNotifications() {
	// In a real implementation, this would schedule periodic checks for inactive accounts
	// For example, it could run daily to check for accounts that will be deleted in 30 days
	log.Println("Scheduling account cleanup notifications...")

	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Run once per day
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// This would call a method to check for inactive accounts and send notifications
				// In a real implementation, this would query the database for accounts that are approaching
				// the 2-year inactivity threshold and send notifications to their owners
				log.Println("Checking for accounts approaching inactivity deletion...")
			}
		}
	}()
}

func (s *EmailService) buildAccountInactivityNotification(userName string) (string, error) {
	tmpl := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Account inactivity notice</title>
		</head>
		<body>
			<h2>Account inactivity notice</h2>
			<p>Hello {{.UserName}},</p>
			<p>This is a courtesy notice that your wish list account has been inactive for an extended period.</p>
			<p>Due to inactivity, your account and associated wish lists will be automatically deleted in 30 days if no activity is detected.</p>
			<p>To prevent deletion, please log in to your account before this period ends.</p>
			<p>If you have any questions, please contact our support team.</p>
			<p>Thank you for using our wish list service.</p>
		</body>
		</html>
	`

	t, err := template.New("accountInactivity").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	data := AccountInactivityNotificationData{
		UserName: userName,
	}

	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *EmailService) SendReservationCancellationEmail(ctx context.Context, recipientEmail, giftItemName, wishlistTitle string) error {
	subject := "Your reservation has been cancelled"
	_, err := s.buildReservationCancellationEmail(giftItemName, wishlistTitle)
	if err != nil {
		return fmt.Errorf("failed to build email body: %w", err)
	}

	// In a real implementation, this would send the email via SMTP
	// Do not log PII (email addresses) or full body content
	log.Printf("Email send simulated: subject=%q (recipient redacted)", subject)

	return nil
}

func (s *EmailService) SendReservationRemovedEmail(ctx context.Context, recipientEmail, giftItemName, wishlistTitle string) error {
	subject := "Your reserved gift item has been removed"
	_, err := s.buildReservationRemovedEmail(giftItemName, wishlistTitle)
	if err != nil {
		return fmt.Errorf("failed to build email body: %w", err)
	}

	// In a real implementation, this would send the email via SMTP
	// Do not log PII (email addresses) or full body content
	log.Printf("Email send simulated: subject=%q (recipient redacted)", subject)

	return nil
}

func (s *EmailService) SendGiftPurchasedConfirmationEmail(ctx context.Context, recipientEmail, giftItemName, wishlistTitle, guestName string) error {
	subject := "Gift Purchased - Thank you!"
	_, err := s.buildGiftPurchasedConfirmationEmail(giftItemName, wishlistTitle, guestName)
	if err != nil {
		return fmt.Errorf("failed to build email body: %w", err)
	}

	// In a real implementation, this would send the email via SMTP
	// Do not log PII (email addresses) or full body content
	log.Printf("Email send simulated: subject=%q (recipient redacted)", subject)

	return nil
}

func (s *EmailService) buildReservationCancellationEmail(giftItemName, wishlistTitle string) (string, error) {
	tmpl := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Your reservation has been cancelled</title>
		</head>
		<body>
			<h2>Your reservation has been cancelled</h2>
			<p>Hello,</p>
			<p>We wanted to inform you that your reservation for the gift item "{{.GiftItemName}}" from the wish list "{{.WishlistTitle}}" has been cancelled.</p>
			<p>If you believe this was done in error, please contact the wish list owner.</p>
			<p>Thank you for using our wish list service.</p>
		</body>
		</html>
	`

	t, err := template.New("reservationCancellation").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	data := ReservationCancellationEmailData{
		GiftItemName:  giftItemName,
		WishlistTitle: wishlistTitle,
	}

	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *EmailService) buildReservationRemovedEmail(giftItemName, wishlistTitle string) (string, error) {
	tmpl := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Your reserved gift item has been removed</title>
		</head>
		<body>
			<h2>Your reserved gift item has been removed</h2>
			<p>Hello,</p>
			<p>We wanted to inform you that the gift item "{{.GiftItemName}}" from the wish list "{{.WishlistTitle}}" that you had reserved has been removed by the wish list owner.</p>
			<p>Your reservation is no longer valid. You may want to consider other gift items on the list.</p>
			<p>Thank you for using our wish list service.</p>
		</body>
		</html>
	`

	t, err := template.New("reservationRemoved").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	data := ReservationRemovedEmailData{
		GiftItemName:  giftItemName,
		WishlistTitle: wishlistTitle,
	}

	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *EmailService) buildGiftPurchasedConfirmationEmail(giftItemName, wishlistTitle, guestName string) (string, error) {
	tmpl := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Gift Purchased - Thank you!</title>
		</head>
		<body>
			<h2>Gift Purchased - Thank you {{.GuestName}}!</h2>
			<p>Hello {{.GuestName}},</p>
			<p>Great news! The wish list owner has confirmed that the gift item "{{.GiftItemName}}" from the wish list "{{.WishlistTitle}}" has been purchased.</p>
			<p>Thank you for your thoughtful gift! The recipient will be delighted.</p>
			<p>Thank you for using our wish list service.</p>
		</body>
		</html>
	`

	t, err := template.New("giftPurchased").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	data := GiftPurchasedConfirmationEmailData{
		GiftItemName:  giftItemName,
		WishlistTitle: wishlistTitle,
		GuestName:     guestName,
	}

	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
