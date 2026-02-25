package notificationsvc

import (
	"context"
	"encoding/json"
	"fmt"

	"firebase.google.com/go/v4/messaging"
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

//	func (ns *NotificationService) SendNotification(ctx context.Context, payload map[string]any) error {
//		return ns.fcmSender.SendPushNotification(ctx, payload)
//	}
func (ns *NotificationService) SendNotification(ctx context.Context, userId string, notif *messaging.Notification, data map[string]string) error {
	// fmt.Println("From Notification service")
	// utils.PrintJSON(data)
	// Handle Notifixation here, FCM, Enail, Text Message, Websocket.
	// For FCM, Get Token by userId,
	const token = "Dummy-Token"
	return ns.fcmSender.SendPushNotification(ctx, token, notif, data)
}

func (ns *NotificationService) SubscribeToInternalTransferCompletedEvents(ctx context.Context) error {
	consumer := utils.WorkerID("notification")

	return ns.eventBus.Subscribe(
		ctx,
		event.StreamTransferEvents,
		"notification-service-group",
		consumer,
		func(ctx context.Context, msg event.EventMessage) error {

			if msg.EventType != event.EventInternalTransferSuccess {
				return nil
			}

			return ns.handleInternalTransferSuccess(ctx, msg)
		},
	)
}

func (ns *NotificationService) handleInternalTransferSuccess(
	ctx context.Context,
	msg event.EventMessage,
) error {

	var evt event.InternalTransferEvent
	if err := json.Unmarshal([]byte(msg.Data), &evt); err != nil {
		return fmt.Errorf("failed to decode %s event: %w",
			event.EventInternalTransferSuccess,
			err,
		)
	}

	// fmt.Println("From HandleInternalTransferSuccess")
	// utils.PrintJSON(evt)

	// Build notification title and body
	notif := &messaging.Notification{
		Title: "Transfer Successful",
		Body:  fmt.Sprintf("You sent ₦%s to %s", evt.Amount, evt.RecipientName),
	}
	// send the data the Client will need
	data := map[string]string{
		"type":          event.EventInternalTransferSuccess,
		"amount":        evt.Amount.String(),
		"recipientName": evt.RecipientName,
		"senderName":    evt.SenderName,
		"toAcctNum":     evt.ToAcctNum,
	}

	// Fire and forget
	_ = ns.SendNotification(ctx, evt.ToAcctUserId.String(), notif, data)

	// Always return nil so event is considered processed
	return nil
}

// package notificationsvc

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"

// 	"github.com/EfosaE/credora-backend/domain/event"
// 	"github.com/EfosaE/credora-backend/domain/fcm"
// 	"github.com/EfosaE/credora-backend/internal/utils"
// )

// type NotificationService struct {
// 	fcmSender fcm.FCMSender
// 	eventBus  event.EventBus
// }

// func NewNotificationService(event event.EventBus, fcm fcm.FCMSender) *NotificationService {
// 	return &NotificationService{
// 		eventBus:  event,
// 		fcmSender: fcm,
// 	}
// }

// func (ns *NotificationService) SendNotification(ctx context.Context, userId string, payload map[string]string) error {
// 	// Handle Notifixation here, FCM, Enail, Text Message, Websocket.
// 	// For FCM, Get Token by userId,
// 	const token = "Dummy-Token"
// 	return ns.fcmSender.SendPushNotification(ctx, token, nil, payload)
// }

// func (ns *NotificationService) handleInternalTransferSuccess(
// 	ctx context.Context,
// 	evt event.MoneyTransferredEvent,
// ) {

// 	notification := &fcm.FCMNotification{
// Title: "Transfer Successful",
// Body:  fmt.Sprintf("You sent ₦%s to %s", evt.Amount, evt.RecipientName),
// 	}

// 	payload := map[string]string{
// 		"type":        "internal_transfer_success",
// 		"amount":      evt.Amount,
// 		"recipientId": evt.RecipientID,
// 	}

// 	err := ns.fcmSender.SendPushNotification(
// 		ctx,
// 		evt.SenderDeviceToken, // get token from event
// 		notification,
// 		payload,
// 	)

// 	if err != nil {
// 		// LOG ONLY. DO NOT RETURN ERROR.
// 		fmt.Printf("push notification failed: %v\n", err)
// 	}
// }

// func (ns *NotificationService) SubscribeToInternalTransferCompletedEvents(ctx context.Context) error {

// 	consumer := utils.WorkerID("notification")

// 	return ns.eventBus.Subscribe(
// 		ctx,
// 		event.StreamTransferEvents,
// 		"notification-service-group",
// 		consumer,
// 		func(ctx context.Context, msg event.EventMessage) error {

// 			if msg.EventType != event.EventInternalTransferSuccess {
// 				return nil
// 			}

// 			var evt event.MoneyTransferredEvent
// 			if err := json.Unmarshal([]byte(msg.Data), &evt); err != nil {
// 				return fmt.Errorf("failed to decode transfer event: %w", err)
// 			}

// 			// DO NOT return error for notification failures so redis doesnt retry
// 			ns.handleInternalTransferSuccess(ctx, evt)

// 			return nil
// 		},
// 	)
// }
