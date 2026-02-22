package fcm

import (
	"context"

	"firebase.google.com/go/v4/messaging"
)

// FCMSender defines the contract for any service that can send
// push notifications via Firebase Cloud Messaging (FCM).
// Implementations could use Firebase Admin SDK or any other FCM-compatible service.
type FCMSender interface {
	// SendPushNotification sends a push notification to a device.
	//
	// Parameters:
	//   ctx    - context for controlling timeout, cancellation, and request scope.
	// 	 token - From Client SDK
	//   fcmN   - pointer to FCMNotification containing the Title and Body to display.
	//   payload - optional custom key-value data to send alongside the notification.
	//
	// Returns:
	//   error - any error encountered while sending the notification.
	SendPushNotification(ctx context.Context, token string, fcmN *messaging.Notification, payload map[string]string) error
}
