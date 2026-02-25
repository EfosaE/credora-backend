package infrastructure

//FIREBASE CLOUD MESSGING FOR SENDING NOTIFICATIONS TO MOBILE APPS
import (
	"context"
	// "fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	// "github.com/EfosaE/credora-backend/internal/utils"
)

type FCMAdapter struct {
	app *firebase.App
	// ADD STRUCTUREE LOGGING
}

func NewFCMAdapter(app *firebase.App) *FCMAdapter {

	return &FCMAdapter{
		app: app,
	}
}

func (f *FCMAdapter) SendPushNotification(
	ctx context.Context,
	token string,
	fcmN *messaging.Notification,
	payload map[string]string,
) error {

	// fmt.Println("SENDING TO MOBILE....")
	// utils.PrintJSON(payload)

	// client, err := f.app.Messaging(ctx)
	// if err != nil {
	// 	fmt.Println("failed to send push notification:", err)
	// 	return fmt.Errorf("failed to get messaging client: %w", err)
	// }

	// message := &messaging.Message{
	// 	Data:  payload,
	// 	Token: token,
	// }

	// // ✅ Add notification only if not nil
	// if fcmN != nil {
	// 	message.Notification = &messaging.Notification{
	// 		Title: fcmN.Title,
	// 		Body:  fcmN.Body,
	// 	}
	// }

	// response, err := client.Send(ctx, message)
	// if err != nil {
	// 	fmt.Println("failed to send push notification:", err)
	// 	return fmt.Errorf("failed to send push notification: %w", err)
	// }

	// fmt.Println("Successfully sent message:", response)
	return nil
}
