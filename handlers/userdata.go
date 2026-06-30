package handlers

import (
	"net/http"

	"github.com/stinkyfingers/poopjournal/auth"
)

func (h *Handler) GetUserDataHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "user not found in context")
		return
	}

	userData, err := h.storage.GetUserData(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user data: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, userData)
}
