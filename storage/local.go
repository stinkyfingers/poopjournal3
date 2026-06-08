package storage

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/stinkyfingers/poopjournal/models"
)

type LocalStorage struct {
	dataDir string
}

func NewLocalStorage(dataDir string) (*LocalStorage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &LocalStorage{
		dataDir: dataDir,
	}, nil
}

func (l *LocalStorage) getUserDir(userID string) string {
	return filepath.Join(l.dataDir, userID)
}

func (l *LocalStorage) SaveFood(ctx context.Context, food *models.Food) error {
	userDir := l.getUserDir(food.UserID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user directory: %w", err)
	}

	foodFile := filepath.Join(userDir, "food.csv")

	// Check if file exists and create header if it doesn't
	var needHeader bool
	if _, err := os.Stat(foodFile); os.IsNotExist(err) {
		needHeader = true
	}

	file, err := os.OpenFile(foodFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open food file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if needHeader {
		header := []string{"id", "user_id", "name", "description", "tags", "timestamp"}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Join tags with semicolon for CSV storage
	tagsStr := ""
	if len(food.Tags) > 0 {
		tagsStr = strings.Join(food.Tags, ";")
	}

	record := []string{
		food.ID,
		food.UserID,
		food.Name,
		food.Description,
		tagsStr,
		food.Timestamp.Format(time.RFC3339),
	}

	return writer.Write(record)
}

func (l *LocalStorage) ListFood(ctx context.Context, userID string) ([]*models.Food, error) {
	foodFile := filepath.Join(l.getUserDir(userID), "food.csv")

	if _, err := os.Stat(foodFile); os.IsNotExist(err) {
		return []*models.Food{}, nil
	}

	file, err := os.Open(foodFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open food file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 { // Header + at least one record
		return []*models.Food{}, nil
	}

	foods := make([]*models.Food, 0, len(records)-1)
	for _, record := range records[1:] { // Skip header
		// Handle both old format (6 fields) and new format (7 fields)
		if len(record) < 6 {
			continue
		}

		// Handle tags for new format
		var tags []string
		timestampIndex := 5

		timestamp, _ := time.Parse(time.RFC3339, record[timestampIndex])

		food := &models.Food{
			ID:          record[0],
			UserID:      record[1],
			Name:        record[2],
			Description: record[3],
			Tags:        tags,
			Timestamp:   timestamp,
		}
		foods = append(foods, food)
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(foods, func(i, j int) bool {
		return foods[i].Timestamp.After(foods[j].Timestamp)
	})

	return foods, nil
}

func (l *LocalStorage) UpdateFood(ctx context.Context, food *models.Food) error {
	foods, err := l.ListFood(ctx, food.UserID)
	if err != nil {
		return err
	}

	// Find and update the food item
	found := false
	for i, f := range foods {
		if f.ID == food.ID {
			foods[i] = food
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("food item not found")
	}

	// Rewrite the entire file
	return l.rewriteFoodFile(food.UserID, foods)
}

func (l *LocalStorage) DeleteFood(ctx context.Context, userID, foodID string) error {
	foods, err := l.ListFood(ctx, userID)
	if err != nil {
		return err
	}

	// Filter out the food item to delete
	filteredFoods := make([]*models.Food, 0, len(foods))
	for _, f := range foods {
		if f.ID != foodID {
			filteredFoods = append(filteredFoods, f)
		}
	}

	return l.rewriteFoodFile(userID, filteredFoods)
}

func (l *LocalStorage) rewriteFoodFile(userID string, foods []*models.Food) error {
	foodFile := filepath.Join(l.getUserDir(userID), "food.csv")

	file, err := os.Create(foodFile)
	if err != nil {
		return fmt.Errorf("failed to create food file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"id", "user_id", "name", "description", "tags", "timestamp"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write records
	for _, food := range foods {
		// Join tags with semicolon for CSV storage
		tagsStr := ""
		if len(food.Tags) > 0 {
			tagsStr = strings.Join(food.Tags, ";")
		}

		record := []string{
			food.ID,
			food.UserID,
			food.Name,
			food.Description,
			tagsStr,
			food.Timestamp.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write food record: %w", err)
		}
	}

	return nil
}

func (l *LocalStorage) SavePoop(ctx context.Context, poop *models.Poop) error {
	userDir := l.getUserDir(poop.UserID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user directory: %w", err)
	}

	poopFile := filepath.Join(userDir, "poop.csv")

	// Check if file exists and create header if it doesn't
	var needHeader bool
	if _, err := os.Stat(poopFile); os.IsNotExist(err) {
		needHeader = true
	}

	file, err := os.OpenFile(poopFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open poop file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if needHeader {
		header := []string{"id", "user_id", "bristol_scale", "urgency", "tags", "notes", "timestamp"}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Join tags with semicolon for CSV storage
	tagsStr := ""
	if len(poop.Tags) > 0 {
		tagsStr = strings.Join(poop.Tags, ";")
	}

	record := []string{
		poop.ID,
		poop.UserID,
		strconv.Itoa(poop.BristolScale),
		strconv.Itoa(poop.Urgency),
		tagsStr,
		poop.Notes,
		poop.Timestamp.Format(time.RFC3339),
	}

	return writer.Write(record)
}

func (l *LocalStorage) ListPoop(ctx context.Context, userID string) ([]*models.Poop, error) {
	poopFile := filepath.Join(l.getUserDir(userID), "poop.csv")

	if _, err := os.Stat(poopFile); os.IsNotExist(err) {
		return []*models.Poop{}, nil
	}

	file, err := os.Open(poopFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open poop file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 { // Header + at least one record
		return []*models.Poop{}, nil
	}

	poops := make([]*models.Poop, 0, len(records)-1)
	for _, record := range records[1:] { // Skip header
		// Handle both old format (6 fields) and new format (8 fields)
		if len(record) < 6 {
			continue
		}

		bristolScale, _ := strconv.Atoi(record[2])

		// Handle urgency and tags for new format
		urgency := 5 // default
		var tags []string
		notesIndex := 5
		timestampIndex := 6

		timestamp, _ := time.Parse(time.RFC3339, record[timestampIndex])

		poop := &models.Poop{
			ID:           record[0],
			UserID:       record[1],
			BristolScale: bristolScale,
			Urgency:      urgency,
			Tags:         tags,
			Notes:        record[notesIndex],
			Timestamp:    timestamp,
		}
		poops = append(poops, poop)
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(poops, func(i, j int) bool {
		return poops[i].Timestamp.After(poops[j].Timestamp)
	})

	return poops, nil
}

func (l *LocalStorage) UpdatePoop(ctx context.Context, poop *models.Poop) error {
	poops, err := l.ListPoop(ctx, poop.UserID)
	if err != nil {
		return err
	}

	// Find and update the poop item
	found := false
	for i, p := range poops {
		if p.ID == poop.ID {
			poops[i] = poop
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("poop item not found")
	}

	// Rewrite the entire file
	return l.rewritePoopFile(poop.UserID, poops)
}

func (l *LocalStorage) DeletePoop(ctx context.Context, userID, poopID string) error {
	poops, err := l.ListPoop(ctx, userID)
	if err != nil {
		return err
	}

	// Filter out the poop item to delete
	filteredPoops := make([]*models.Poop, 0, len(poops))
	for _, p := range poops {
		if p.ID != poopID {
			filteredPoops = append(filteredPoops, p)
		}
	}

	return l.rewritePoopFile(userID, filteredPoops)
}

func (l *LocalStorage) rewritePoopFile(userID string, poops []*models.Poop) error {
	poopFile := filepath.Join(l.getUserDir(userID), "poop.csv")

	file, err := os.Create(poopFile)
	if err != nil {
		return fmt.Errorf("failed to create poop file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"id", "user_id", "bristol_scale", "urgency", "tags", "notes", "timestamp"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write records
	for _, poop := range poops {
		// Join tags with semicolon for CSV storage
		tagsStr := ""
		if len(poop.Tags) > 0 {
			tagsStr = strings.Join(poop.Tags, ";")
		}

		record := []string{
			poop.ID,
			poop.UserID,
			strconv.Itoa(poop.BristolScale),
			strconv.Itoa(poop.Urgency),
			tagsStr,
			poop.Notes,
			poop.Timestamp.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write poop record: %w", err)
		}
	}

	return nil
}

func (l *LocalStorage) GetAllPoopTags(ctx context.Context, userID string) ([]string, error) {
	poops, err := l.ListPoop(ctx, userID)
	if err != nil {
		return nil, err
	}

	tagSet := make(map[string]bool)
	for _, poop := range poops {
		for _, tag := range poop.Tags {
			if tag != "" {
				tagSet[tag] = true
			}
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	sort.Strings(tags)
	return tags, nil
}

func (l *LocalStorage) GetAllFoodTags(ctx context.Context, userID string) ([]string, error) {
	foods, err := l.ListFood(ctx, userID)
	if err != nil {
		return nil, err
	}

	tagSet := make(map[string]bool)
	// Always include 'medicine' as an option
	tagSet["medicine"] = true

	for _, food := range foods {
		for _, tag := range food.Tags {
			if tag != "" {
				tagSet[tag] = true
			}
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	sort.Strings(tags)
	return tags, nil
}
