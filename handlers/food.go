package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

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

type foodMutationRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	DateTime    string   `json:"datetime"`
	Tags        []string `json:"tags"`
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

	req, err := parseFoodMutationRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "food name is required")
		return
	}

	if req.DateTime == "" {
		writeError(w, http.StatusBadRequest, "date and time is required")
		return
	}

	food := models.NewFoodWithDateTime(userId, req.Name, req.Description, req.DateTime, req.Tags)
	if err := h.storage.SaveFood(r.Context(), food); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save food entry")
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

	path := strings.TrimPrefix(r.URL.Path, "/food/")
	foodID := strings.Split(path, "/")[0]

	req, err := parseFoodMutationRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "food name is required")
		return
	}

	if req.DateTime == "" {
		writeError(w, http.StatusBadRequest, "date and time is required")
		return
	}

	foods, err := h.storage.ListFood(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get food entries")
		return
	}

	var food *models.Food
	for _, f := range foods {
		if f.ID == foodID {
			food = f
			break
		}
	}

	if food == nil {
		writeError(w, http.StatusNotFound, "food entry not found")
		return
	}

	timestamp, err := time.Parse("2006-01-02T15:04", req.DateTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date and time format")
		return
	}

	food.Name = req.Name
	food.Description = req.Description
	food.Timestamp = timestamp
	food.Tags = normalizeTags(req.Tags)

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

	path := strings.TrimPrefix(r.URL.Path, "/food/")
	foodID := strings.Split(path, "/")[0]

	if err := h.storage.DeleteFood(r.Context(), userId, foodID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete food entry")
		return
	}

	writeJSON(w, http.StatusOK, deleteResponse{Deleted: true, ID: foodID})
}

func parseFoodMutationRequest(r *http.Request) (*foodMutationRequest, error) {
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		defer r.Body.Close()

		var req foodMutationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if errors.Is(err, io.EOF) {
				return nil, errors.New("request body is required")
			}
			return nil, errors.New("failed to decode JSON body")
		}

		req.Name = strings.TrimSpace(req.Name)
		req.Description = strings.TrimSpace(req.Description)
		req.DateTime = strings.TrimSpace(req.DateTime)
		req.Tags = normalizeTags(req.Tags)
		return &req, nil
	}

	if err := r.ParseForm(); err != nil {
		return nil, errors.New("failed to parse form body")
	}

	return &foodMutationRequest{
		Name:        strings.TrimSpace(r.FormValue("name")),
		Description: strings.TrimSpace(r.FormValue("description")),
		DateTime:    strings.TrimSpace(r.FormValue("datetime")),
		Tags:        normalizeTags(strings.Split(strings.TrimSpace(r.FormValue("tags")), ",")),
	}, nil
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
