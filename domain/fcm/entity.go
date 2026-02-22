package fcm

// SendNotificationRequest represents a simplified request payload
// for sending a push notification to a single device.
// Can include either a Notification (Title + Body) or custom Data payload.
type SendNotificationRequest struct {
	Token string            `json:"token"`           // FCM device token to send the message to
	Title string            `json:"title,omitempty"` // Notification title (optional)
	Body  string            `json:"body,omitempty"`  // Notification body/content (optional)
	Data  map[string]string `json:"data,omitempty"`  // Custom key-value data payload (optional)
}

// FCMRequest represents the full structure expected by Firebase Cloud Messaging.
// This mirrors the JSON structure Firebase Admin SDK can use.
type FCMRequest struct {
	Message FCMMessage `json:"message"` // The actual message to send
}

// FCMMessage represents a single FCM message.
// Can contain a notification, data payload, or both.
type FCMMessage struct {
	Token        string            `json:"token"`                  // FCM device token
	Notification *FCMNotification  `json:"notification,omitempty"` // Notification content (optional)
	Data         map[string]string `json:"data,omitempty"`         // Custom key-value data payload (optional)
}

// FCMNotification represents the notification part of a message.
// This is the part displayed by the OS in the system tray.
type FCMNotification struct {
	Title string `json:"title"` // Notification title shown to the user
	Body  string `json:"body"`  // Notification message body/content shown to the user
}
