package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/stinkyfingers/poopjournal/auth"
	"github.com/stinkyfingers/poopjournal/models"
)

func (h *Handler) AddFoodHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "user not found in context")
		return
	}

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var food models.Food
	err := json.NewDecoder(r.Body).Decode(&food)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to decode JSON body: "+err.Error())
		return
	}

	if food.Name == "" {
		writeError(w, http.StatusBadRequest, "food name is required")
		return
	}

	if food.Timestamp.IsZero() {
		writeError(w, http.StatusBadRequest, "date and time is required")
		return
	}
	userData, err := h.storage.GetUserData(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	userData.Foods = append(userData.Foods, food)

	if err := h.storage.SaveUserData(r.Context(), userId, userData); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save food entry: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, food)
}

func (h *Handler) UpdateFoodHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "user not found in context")
		return
	}

	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var food models.Food
	err := json.NewDecoder(r.Body).Decode(&food)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to decode JSON body: "+err.Error())
		return
	}
	if food.ID == "" {
		writeError(w, http.StatusBadRequest, "food ID is required for update")
		return
	}
	if food.Name == "" {
		writeError(w, http.StatusBadRequest, "food name is required")
		return
	}

	if food.Timestamp.IsZero() {
		writeError(w, http.StatusBadRequest, "date and time is required")
		return
	}
	food.UserID = userId
	food.Tags = normalizeTags(food.Tags)

	userData, err := h.storage.GetUserData(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i, userFood := range userData.Foods {
		if food.ID == userFood.ID {
			userData.Foods[i] = food
		}
	}

	if err := h.storage.SaveUserData(r.Context(), userId, userData); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save food entry: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, food)
}

func (h *Handler) DeleteFoodHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "user not found in context")
		return
	}

	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	foodID := r.URL.Query().Get("id")
	if foodID == "" {
		writeError(w, http.StatusBadRequest, "food ID is required for deletion")
		return
	}

	userData, err := h.storage.GetUserData(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i, userFood := range userData.Foods {
		if foodID == userFood.ID {
			userData.Foods = append(userData.Foods[:i], userData.Foods[i+1:]...)
		}
	}

	if err := h.storage.SaveUserData(r.Context(), userId, userData); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save food entry: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, deleteResponse{Deleted: true, ID: foodID})
}

func normalizeTags(tags []string) []string {
	validatedTags := make([]string, 0)
	if len(tags) > 0 {
		seen := make(map[string]bool)
		for _, tag := range tags {
			cleaned := strings.ToLower(strings.TrimSpace(tag))
			if cleaned != "" && len(cleaned) <= 20 && !seen[cleaned] && len(validatedTags) < 5 {
				seen[cleaned] = true
				validatedTags = append(validatedTags, cleaned)
			}
		}
	}
	return validatedTags
}
