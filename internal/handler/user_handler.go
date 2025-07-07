package handler



// func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
// 	idStr := chi.URLParam(r, "id")
// 	id, err := strconv.ParseInt(idStr, 10, 64)
// 	if err != nil {
// 		http.Error(w, "invalid ID", http.StatusBadRequest)
// 		return
// 	}

// 	user, err := h.userService.GetUserByID(id)
// 	if err != nil {
// 		http.Error(w, "user not found", http.StatusNotFound)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(user)
// }