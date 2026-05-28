package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/stinkyfingers/poopjournal/auth"
	"github.com/stinkyfingers/poopjournal/models"
	"github.com/stinkyfingers/poopjournal/storage"
)

type FoodHandler struct {
	storage storage.Storage
}

func NewFoodHandler(storage storage.Storage) *FoodHandler {
	return &FoodHandler{
		storage: storage,
	}
}

type foodPageResponse struct {
	Foods        []*models.Food `json:"foods"`
	ExistingTags []string       `json:"existing_tags"`
}

type deleteResponse struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

func (h *FoodHandler) ListFoodHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "user not found in context")
		return
	}

	foods, err := h.storage.ListFood(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get food entries")
		return
	}

	existingTags, err := h.storage.GetAllFoodTags(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get existing tags")
		return
	}

	sort.Strings(existingTags)
	writeJSON(w, http.StatusOK, foodPageResponse{Foods: foods, ExistingTags: existingTags})
}

func (h *FoodHandler) AddFoodHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "user not found in context")
		return
	}

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var food *models.Food
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
	food.ID = uuid.New().String()
	food.UserID = userId
	food.Tags = normalizeTags(food.Tags)
	if err := h.storage.SaveFood(r.Context(), food); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add food entry")
		return
	}

	writeJSON(w, http.StatusCreated, food)
}

func (h *FoodHandler) UpdateFoodHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "user not found in context")
		return
	}

	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var food *models.Food
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

	if err := h.storage.UpdateFood(r.Context(), food); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update food entry")
		return
	}

	writeJSON(w, http.StatusOK, food)
}

func (h *FoodHandler) DeleteFoodHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := h.storage.DeleteFood(r.Context(), userId, foodID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete food entry")
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
