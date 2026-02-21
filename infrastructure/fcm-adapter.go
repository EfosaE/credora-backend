package infrastructure

//FIREBASE CLOUD MESSGING FOR SENDING NOTIFICATIONS TO MOBILE APPS

import (
	"context"
	"fmt"

	"github.com/EfosaE/credora-backend/internal/utils"
)

type fcmConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}
type FCMAdapter struct {
	fcmConfig
}

func NewFCMAdapter() *FCMAdapter {
	// port, err := strconv.Atoi(config.App.MailtrapPort)
	// if err != nil {
	// 	log.Fatal("Invalid port:", err)
	// }
	return &FCMAdapter{
		// fcmConfig: fcmConfig{
		// 	Host:     config.App.MailtrapHost,
		// 	Port:     port,
		// 	Username: config.App.MailtrapUser,
		// 	Password: config.App.MailtrapPass,
		// },
	}
}

func (s *FCMAdapter) SendPushNotification(ctx context.Context, payload map[string]any) error {
	fmt.Println("SENDING TO MOBILE....")
	utils.PrintJSON(payload)
	return nil
}
