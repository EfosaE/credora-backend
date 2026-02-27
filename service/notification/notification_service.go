package notificationsvc

import (
	"context"
	"encoding/json"
	"fmt"

	"firebase.google.com/go/v4/messaging"
	"github.com/EfosaE/credora-backend/domain/account"
	"github.com/EfosaE/credora-backend/domain/event"
	"github.com/EfosaE/credora-backend/domain/fcm"
	"github.com/EfosaE/credora-backend/domain/user"
	"github.com/EfosaE/credora-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type NotificationService struct {
	fcmSender  fcm.FCMSender
	eventBus   event.EventBus
	acctRepo   account.AccountRepository
	deviceRepo user.DeviceRepository
	logger     zerolog.Logger
}

func NewNotificationService(
	e event.EventBus,
	fcm fcm.FCMSender,
	a account.AccountRepository,
	d user.DeviceRepository,
	l zerolog.Logger,
) *NotificationService {

	logger := l.With().
		Str("service", "notification-service").
		Logger()

	logger.Info().Msg("notification service initialized")

	return &NotificationService{
		eventBus:   e,
		fcmSender:  fcm,
		acctRepo:   a,
		deviceRepo: d,
		logger:     logger,
	}
}

func (ns *NotificationService) SendNotification(
	ctx context.Context,
	userId uuid.UUID,
	notif *messaging.Notification,
	data map[string]string,
) error {

	log := ns.logger.With().
		Str("operation", "SendNotification").
		Str("user_id", userId.String()).
		Logger()

	tokens, err := ns.deviceRepo.GetByUserID(ctx, userId)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to fetch device tokens")
		return err
	}

	if len(tokens) == 0 {
		log.Info().
			Msg("no registered device tokens found for user")
		return nil
	}

	log.Info().
		Int("device_count", len(tokens)).
		Msg("sending push notification to devices")

	for _, device := range tokens {

		err := ns.fcmSender.SendPushNotification(
			ctx,
			device.Token,
			notif,
			data,
		)

		if err != nil {
			log.Error().
				Str("token", device.Token).
				Err(err).
				Msg("failed to send push notification")
			continue
		}

		log.Info().
			Str("token", device.Token).
			Msg("push notification sent successfully")
	}

	return nil
}

func (ns *NotificationService) SubscribeToInternalTransferCompletedEvents(ctx context.Context) error {

	consumer := utils.WorkerID("notification")

	ns.logger.Info().
		Str("stream", event.StreamTransferEvents).
		Str("consumer_group", "notification-service-group").
		Str("consumer_id", consumer).
		Msg("subscribing to internal transfer events")

	return ns.eventBus.Subscribe(
		ctx,
		event.StreamTransferEvents,
		"notification-service-group",
		consumer,
		func(ctx context.Context, msg event.EventMessage) error {

			ns.logger.Info().
				Str("event_type", msg.EventType).
				Msg("event received")

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

	log := ns.logger.With().
		Str("operation", "handleInternalTransferSuccess").
		Str("event_type", msg.EventType).
		Logger()

	var evt event.InternalTransferEvent
	if err := json.Unmarshal([]byte(msg.Data), &evt); err != nil {
		log.Error().
			Err(err).
			Str("raw_data", msg.Data).
			Msg("failed to decode internal transfer event")
		return fmt.Errorf("failed to decode %s event: %w",
			event.EventInternalTransferSuccess,
			err,
		)
	}

	log.Info().
		Str("from_account", evt.FromAcctNum).
		Str("to_account", evt.ToAcctNum).
		Str("amount", evt.Amount.String()).
		Msg("processing internal transfer success event")

	accounts, err := ns.acctRepo.GetAccountsDetails(
		ctx,
		[]string{evt.FromAcctNum, evt.ToAcctNum},
	)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to fetch account details")
		return err
	}

	accountMap := make(map[string]*account.Account)
	for _, acc := range accounts {
		accountMap[acc.AccountNumber] = acc
	}

	fromAcct, ok := accountMap[evt.FromAcctNum]
	if !ok {
		log.Error().
			Str("account_number", evt.FromAcctNum).
			Msg("sender account not found")
		return fmt.Errorf("from account %s not found", evt.FromAcctNum)
	}

	toAcct, ok := accountMap[evt.ToAcctNum]
	if !ok {
		log.Error().
			Str("account_number", evt.ToAcctNum).
			Msg("recipient account not found")
		return fmt.Errorf("to account %s not found", evt.ToAcctNum)
	}

	amountStr := evt.Amount.StringFixed(2)

	notif := &messaging.Notification{
		Title: "Credit Alert",
		Body:  fmt.Sprintf("₦%s received from %s", amountStr, fromAcct.UserName),
	}

	data := map[string]string{
		"type":          event.EventInternalTransferSuccess,
		"amount":        evt.Amount.String(),
		"recipientName": toAcct.UserName,
		"senderName":    fromAcct.UserName,
		"toAcctNum":     evt.ToAcctNum,
	}

	log.Info().
		Str("recipient_user_id", evt.ToAcctUserId.String()).
		Msg("dispatching credit notification")

	err = ns.SendNotification(ctx, evt.ToAcctUserId, notif, data)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to dispatch notification")
		return err
	}

	log.Info().
		Msg("internal transfer notification processed successfully")

	return nil
}
