package router

import (
	"github.com/EfosaE/credora-backend/internal/handler"
	"github.com/go-chi/chi/v5"
)

func RegisterMonnifyRoutes(r chi.Router, h *handler.MonnifyHandler) {
	r.Route("/reserved-account", func(r chi.Router) {
		r.Delete("/{acctRef}", h.DeleteCustomerHandler)
	})
}
