// handlers/user_handler.go
package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/EfosaE/credora-backend/service"
	"github.com/go-chi/chi/v5"
)

type MonnifyHandler struct {
	monnifyService *service.MonnifyService
}

func NewMonnifyHandler(monnifySvc *service.MonnifyService) *MonnifyHandler {
	return &MonnifyHandler{monnifyService: monnifySvc}
}

// DeleteCustomerHandler handles deleting a reserved Monnify account by account reference (acctRef)
func (h *MonnifyHandler) DeleteCustomerHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("DEBUG acctRef param:", chi.URLParam(r, "acctRef"))

	fmt.Println("Full URL:", r.URL.Path)
	fmt.Println("acctRef param:", chi.URLParam(r, "acctRef"))
	// Extract acctRef from the URL
	acctRef := chi.URLParam(r, "acctRef")
	if strings.TrimSpace(acctRef) == "" {
		response.SendError(w, r, response.BadRequest(nil, "account reference is required"))
		return
	}

	// Call the Monnify service to delete the reserved account
	result, err := h.monnifyService.DeleteCustomer(acctRef)
	if err != nil {
		fmt.Println("Monnify delete error:", err)
		response.SendError(w, r, response.InternalServerError(err, "could not delete reserved account"))
		return
	}

	// Success
	response.SendSuccess(w, r, response.OK(result, "Reserved account deleted successfully"))
}
