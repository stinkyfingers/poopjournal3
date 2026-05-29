package handlers

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/google/uuid"
	"github.com/stinkyfingers/poopjournal/auth"
	"github.com/stinkyfingers/poopjournal/models"
	"github.com/stinkyfingers/poopjournal/storage"
)

type PoopHandler struct {
	storage storage.Storage
}

func NewPoopHandler(storage storage.Storage) *PoopHandler {
	return &PoopHandler{
		storage: storage,
	}
}

type poopPageResponse struct {
	Poops        []*models.Poop `json:"poops"`
	ExistingTags []string       `json:"existing_tags"`
}

func (h *PoopHandler) ListPoopHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "user not found in context")
		return
	}

	poops, err := h.storage.ListPoop(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get poop entries: "+err.Error())
		return
	}

	existingTags, err := h.storage.GetAllPoopTags(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get existing tags: "+err.Error())
		return
	}

	sort.Strings(existingTags)
	writeJSON(w, http.StatusOK, poopPageResponse{Poops: poops, ExistingTags: existingTags})
}

func (h *PoopHandler) AddPoopHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := h.storage.SavePoop(r.Context(), &poop); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save poop entry: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, poop)
}

func (h *PoopHandler) UpdatePoopHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := h.storage.UpdatePoop(r.Context(), &poop); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update poop entry: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, poop)
}

func (h *PoopHandler) DeletePoopHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := h.storage.DeletePoop(r.Context(), userId, poopID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete poop entry: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, deleteResponse{Deleted: true, ID: poopID})
}
