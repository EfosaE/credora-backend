package notificationsvc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/fcm"
	"github.com/EfosaE/credora-backend/internal/utils"
)

type NotificationService struct {
	fcmSender fcm.FCMSender
	eventBus  event.EventBus
}

func NewNotificationService(event event.EventBus, fcm fcm.FCMSender) *NotificationService {
	return &NotificationService{
		eventBus:  event,
		fcmSender: fcm,
	}
}

func (ns *NotificationService) SendPushNotification(ctx context.Context, payload map[string]any) error {
	return ns.fcmSender.SendPushNotification(ctx, payload)
}

func (ns *NotificationService) SubscribeToInternalTransferCompletedEvents(ctx context.Context) error {
	consumer := utils.WorkerID("notification")
	return ns.eventBus.Subscribe(
		ctx,
		event.StreamTransferEvents,
		"notification-service-group",
		consumer,
		func(ctx context.Context, msg event.EventMessage) error {

			// Ignore other event types if stream is shared
			if msg.EventType != event.EventInternalTransferSuccess {
				return nil
			}

			var evt event.MoneyTransferredEvent
			if err := json.Unmarshal([]byte(msg.Data), &evt); err != nil {
				return fmt.Errorf("failed to decode %s event: %w", event.EventUserCreated, err)
			}

			payload, err := utils.StructToMap(evt)
			if err != nil {
				return err
			}
			// Send email REMEMBER TO ENQUEUE THIS
			if err := ns.SendPushNotification(
				ctx,
				payload,
			); err != nil {
				return fmt.Errorf("failed to send push notification: %w", err)
			}

			return nil
		},
	)
}
