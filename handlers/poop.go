package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

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

type poopMutationRequest struct {
	BristolScale int      `json:"bristol_scale"`
	Urgency      int      `json:"urgency"`
	Notes        string   `json:"notes"`
	DateTime     string   `json:"datetime"`
	Tags         []string `json:"tags"`
}

func (h *PoopHandler) ListPoopHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "user not found in context")
		return
	}

	poops, err := h.storage.ListPoop(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get poop entries")
		return
	}

	existingTags, err := h.storage.GetAllPoopTags(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get existing tags")
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

	req, err := parsePoopMutationRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.BristolScale < 1 || req.BristolScale > 7 {
		writeError(w, http.StatusBadRequest, "invalid Bristol scale value (must be 1-7)")
		return
	}

	if req.Urgency < 1 || req.Urgency > 10 {
		writeError(w, http.StatusBadRequest, "invalid urgency value (must be 1-10)")
		return
	}

	if req.DateTime == "" {
		writeError(w, http.StatusBadRequest, "date and time is required")
		return
	}

	poop := models.NewPoopWithDateTime(userId, req.BristolScale, req.Urgency, req.Notes, req.DateTime, req.Tags)
	if err := h.storage.SavePoop(r.Context(), poop); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save poop entry")
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

	path := strings.TrimPrefix(r.URL.Path, "/poop/")
	poopID := strings.Split(path, "/")[0]

	req, err := parsePoopMutationRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.BristolScale < 1 || req.BristolScale > 7 {
		writeError(w, http.StatusBadRequest, "invalid Bristol scale value (must be 1-7)")
		return
	}

	if req.Urgency < 1 || req.Urgency > 10 {
		writeError(w, http.StatusBadRequest, "invalid urgency value (must be 1-10)")
		return
	}

	if req.DateTime == "" {
		writeError(w, http.StatusBadRequest, "date and time is required")
		return
	}

	poops, err := h.storage.ListPoop(r.Context(), userId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get poop entries")
		return
	}

	var poop *models.Poop
	for _, p := range poops {
		if p.ID == poopID {
			poop = p
			break
		}
	}

	if poop == nil {
		writeError(w, http.StatusNotFound, "poop entry not found")
		return
	}

	timestamp, err := time.Parse("2006-01-02T15:04", req.DateTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date and time format")
		return
	}

	poop.BristolScale = req.BristolScale
	poop.Urgency = req.Urgency
	poop.Notes = req.Notes
	poop.Timestamp = timestamp
	poop.Tags = normalizeTags(req.Tags)

	if err := h.storage.UpdatePoop(r.Context(), poop); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update poop entry")
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

	path := strings.TrimPrefix(r.URL.Path, "/poop/")
	poopID := strings.Split(path, "/")[0]

	if err := h.storage.DeletePoop(r.Context(), userId, poopID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete poop entry")
		return
	}

	writeJSON(w, http.StatusOK, deleteResponse{Deleted: true, ID: poopID})
}

func parsePoopMutationRequest(r *http.Request) (*poopMutationRequest, error) {
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		defer r.Body.Close()

		var req poopMutationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if errors.Is(err, io.EOF) {
				return nil, errors.New("request body is required")
			}
			return nil, errors.New("failed to decode JSON body")
		}

		req.Notes = strings.TrimSpace(req.Notes)
		req.DateTime = strings.TrimSpace(req.DateTime)
		req.Tags = normalizeTags(req.Tags)
		return &req, nil
	}

	if err := r.ParseForm(); err != nil {
		return nil, errors.New("failed to parse form body")
	}

	bristolScale, err := strconv.Atoi(r.FormValue("bristol_scale"))
	if err != nil {
		bristolScale = 0
	}

	urgency, err := strconv.Atoi(r.FormValue("urgency"))
	if err != nil {
		urgency = 0
	}

	return &poopMutationRequest{
		BristolScale: bristolScale,
		Urgency:      urgency,
		Notes:        strings.TrimSpace(r.FormValue("notes")),
		DateTime:     strings.TrimSpace(r.FormValue("datetime")),
		Tags:         normalizeTags(strings.Split(strings.TrimSpace(r.FormValue("tags")), ",")),
	}, nil
}

func getBristolDescription(scale int) string {
	descriptions := map[int]string{
		1: "Hard lumps",
		2: "Sausage-shaped but lumpy",
		3: "Sausage-shaped with cracks",
		4: "Smooth and soft sausage",
		5: "Soft blobs with clear edges",
		6: "Fluffy pieces with ragged edges",
		7: "Watery, no solid pieces",
	}

	if desc, ok := descriptions[scale]; ok {
		return desc
	}
	return "Unknown"
}
