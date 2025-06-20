package router

import (
	"github.com/EfosaE/credora-backend/internal/handler"
	"github.com/go-chi/chi/v5"
)

func RegisterUserRoutes(r chi.Router, h *handler.UserHandler) {
    r.Route("/users", func(r chi.Router) {
        r.Post("/", h.CreateUserHandler)
        // r.Get("/{id}", h.GetUserHandler)
    })
}
