package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/stinkyfingers/poopjournal/auth"
	"github.com/stinkyfingers/poopjournal/models"
)

func (h *Handler) AddPoopHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "user not found in context")
		return
	}

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var poop models.Poop
	err := json.NewDecoder(r.Body).Decode(&poop)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to decode JSON body: "+err.Error())
		return
	}

	if poop.BristolScale < 1 || poop.BristolScale > 7 {
		writeError(w, http.StatusBadRequest, "invalid Bristol scale value (must be 1-7)")
		return
	}

	if poop.Urgency < 1 || poop.Urgency > 10 {
		writeError(w, http.StatusBadRequest, "invalid urgency value (must be 1-10)")
		return
	}

	if poop.Timestamp.IsZero() {
		writeError(w, http.StatusBadRequest, "date and time is required")
		return
	}
	poop.ID = uuid.New().String()
	poop.UserID = userId
	poop.Tags = normalizeTags(poop.Tags)

	userData, err := h.storage.GetUserData(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	userData.Poops = append(userData.Poops, poop)

	if err := h.storage.SaveUserData(r.Context(), userId, userData); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save poop entry: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, poop)
}

func (h *Handler) UpdatePoopHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "user not found in context")
		return
	}

	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var poop models.Poop
	err := json.NewDecoder(r.Body).Decode(&poop)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to decode JSON body: "+err.Error())
		return
	}

	if poop.BristolScale < 1 || poop.BristolScale > 7 {
		writeError(w, http.StatusBadRequest, "invalid Bristol scale value (must be 1-7)")
		return
	}

	if poop.Urgency < 1 || poop.Urgency > 10 {
		writeError(w, http.StatusBadRequest, "invalid urgency value (must be 1-10)")
		return
	}

	if poop.Timestamp.IsZero() {
		writeError(w, http.StatusBadRequest, "date and time is required")
		return
	}
	if poop.ID == "" {
		writeError(w, http.StatusBadRequest, "poop ID is required for update")
		return
	}

	poop.UserID = userId
	poop.Tags = normalizeTags(poop.Tags)

	userData, err := h.storage.GetUserData(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i, userPoop := range userData.Poops {
		if poop.ID == userPoop.ID {
			userData.Poops[i] = poop
		}
	}

	if err := h.storage.SaveUserData(r.Context(), userId, userData); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save poop entry: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, poop)
}

func (h *Handler) DeletePoopHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "user not found in context")
		return
	}

	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	poopID := r.URL.Query().Get("id")
	if poopID == "" {
		writeError(w, http.StatusBadRequest, "poop ID is required for deletion")
		return
	}

	userData, err := h.storage.GetUserData(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i, userPoop := range userData.Poops {
		if poopID == userPoop.ID {
			userData.Poops = append(userData.Poops[:i], userData.Poops[i+1:]...)
		}
	}

	if err := h.storage.SaveUserData(r.Context(), userId, userData); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save poop entry: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, deleteResponse{Deleted: true, ID: poopID})
}
