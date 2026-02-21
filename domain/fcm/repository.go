package fcm

import (
	"context"
)

type FCMSender interface {
	SendPushNotification(ctx context.Context, payload map[string]any) error
}

type FCMPayload struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Subject  string `json:"subject"`
	Html     string `json:"html,omitempty"`     // raw HTML if already built
	Template string `json:"template,omitempty"` // template name, if using embedded templates
	Data     any    `json:"data,omitempty"`     // template data context
	ReplyTo  string `json:"replyTo,omitempty"`
}
