// handlers/user_handler.go
package handler

import (
	"fmt"
	"net/http"

	"firebase.google.com/go/v4/messaging"
	"github.com/EfosaE/credora-backend/internal/response"
	notificationsvc "github.com/EfosaE/credora-backend/service/notification"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	notifSvc *notificationsvc.NotificationService
}

func NewNotificationHandler(notifSvc *notificationsvc.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifSvc: notifSvc}
}

func (h *NotificationHandler) SendNotificationHandler(w http.ResponseWriter, r *http.Request) {

	_, claims, _ := jwtauth.FromContext(r.Context())

	userIDStr, ok := claims["userId"].(string)
	if !ok {
		response.SendError(w, r, response.Unauthorized("invalid userId in token"))
		return
	}

	userID, iErr := uuid.Parse(userIDStr)
	if iErr != nil {
		response.SendError(w, r, response.Unauthorized("invalid uuid format"))
		return
	}
	fcmN := &messaging.Notification{
		Title: "Hello from the server",
		Body:  "This is a test notifcation",
	}

	// Call the Monnify service to delete the reserved account
	err := h.notifSvc.SendNotification(r.Context(), userID, fcmN, nil)
	if err != nil {
		fmt.Println("Error occurred in sending notification", err)
		response.SendError(w, r, response.InternalServerError(err, "failed to send notifcation"))
		return
	}

	// Success
	response.SendSuccess(w, r, response.OK(
		nil,
		nil,
		"Notification sent, check your client app",
	))
}
