package storage

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/stinkyfingers/poopjournal/models"
)

type S3Storage struct {
	client     *s3.Client
	bucketName string
}

func NewS3Storage(bucketName string) (*S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Storage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (s *S3Storage) getUserKey(userID, fileName string) string {
	// Sanitize userID for use as S3 key (if needed)
	sanitized := strings.ReplaceAll(userID, "@", "_at_")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	return fmt.Sprintf("users/%s/%s", sanitized, fileName)
}

func (s *S3Storage) SaveFood(ctx context.Context, food *models.Food) error {
	key := s.getUserKey(food.UserID, "food.csv")

	// Get existing foods
	foods, err := s.ListFood(ctx, food.UserID)
	if err != nil {
		return fmt.Errorf("failed to get existing foods: %w", err)
	}

	// Add the new food
	foods = append([]*models.Food{food}, foods...)

	// Write back to S3
	return s.writeFoodFile(ctx, key, foods)
}

func (s *S3Storage) ListFood(ctx context.Context, userID string) ([]*models.Food, error) {
	key := s.getUserKey(userID, "food.csv")

	// Try to get the file from S3
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			// File doesn't exist, return empty slice
			return []*models.Food{}, nil
		}
		return nil, fmt.Errorf("failed to get food file from S3: %w", err)
	}
	defer result.Body.Close()

	// Read and parse CSV
	reader := csv.NewReader(result.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 { // Header + at least one record
		return []*models.Food{}, nil
	}

	foods := make([]*models.Food, 0, len(records)-1)
	for _, record := range records[1:] { // Skip header
		if len(record) < 6 {
			continue
		}

		timestamp, _ := time.Parse(time.RFC3339, record[4])

		food := &models.Food{
			ID:          record[0],
			UserID:      record[1],
			Name:        record[2],
			Description: record[3],
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

func (s *S3Storage) UpdateFood(ctx context.Context, food *models.Food) error {
	foods, err := s.ListFood(ctx, food.UserID)
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

	key := s.getUserKey(food.UserID, "food.csv")
	return s.writeFoodFile(ctx, key, foods)
}

func (s *S3Storage) DeleteFood(ctx context.Context, userID, foodID string) error {
	foods, err := s.ListFood(ctx, userID)
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

	key := s.getUserKey(userID, "food.csv")
	return s.writeFoodFile(ctx, key, filteredFoods)
}

func (s *S3Storage) writeFoodFile(ctx context.Context, key string, foods []*models.Food) error {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"id", "user_id", "name", "description", "timestamp", "created_at"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write records
	for _, food := range foods {
		record := []string{
			food.ID,
			food.UserID,
			food.Name,
			food.Description,
			food.Timestamp.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write food record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	// Upload to S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("text/csv"),
	})

	return err
}

func (s *S3Storage) SavePoop(ctx context.Context, poop *models.Poop) error {
	key := s.getUserKey(poop.UserID, "poop.csv")

	// Get existing poops
	poops, err := s.ListPoop(ctx, poop.UserID)
	if err != nil {
		return fmt.Errorf("failed to get existing poops: %w", err)
	}

	// Add the new poop
	poops = append([]*models.Poop{poop}, poops...)

	// Write back to S3
	return s.writePoopFile(ctx, key, poops)
}

func (s *S3Storage) ListPoop(ctx context.Context, userID string) ([]*models.Poop, error) {
	key := s.getUserKey(userID, "poop.csv")

	// Try to get the file from S3
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			// File doesn't exist, return empty slice
			return []*models.Poop{}, nil
		}
		return nil, fmt.Errorf("failed to get poop file from S3: %w", err)
	}
	defer result.Body.Close()

	// Read and parse CSV
	reader := csv.NewReader(result.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 { // Header + at least one record
		return []*models.Poop{}, nil
	}

	poops := make([]*models.Poop, 0, len(records)-1)
	for _, record := range records[1:] { // Skip header
		if len(record) < 6 {
			continue
		}

		bristolScale, _ := strconv.Atoi(record[2])
		timestamp, _ := time.Parse(time.RFC3339, record[4])

		poop := &models.Poop{
			ID:           record[0],
			UserID:       record[1],
			BristolScale: bristolScale,
			Notes:        record[3],
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

func (s *S3Storage) UpdatePoop(ctx context.Context, poop *models.Poop) error {
	poops, err := s.ListPoop(ctx, poop.UserID)
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

	key := s.getUserKey(poop.UserID, "poop.csv")
	return s.writePoopFile(ctx, key, poops)
}

func (s *S3Storage) DeletePoop(ctx context.Context, userID, poopID string) error {
	poops, err := s.ListPoop(ctx, userID)
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

	key := s.getUserKey(userID, "poop.csv")
	return s.writePoopFile(ctx, key, filteredPoops)
}

func (s *S3Storage) writePoopFile(ctx context.Context, key string, poops []*models.Poop) error {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"id", "user_id", "bristol_scale", "notes", "timestamp", "created_at"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write records
	for _, poop := range poops {
		record := []string{
			poop.ID,
			poop.UserID,
			strconv.Itoa(poop.BristolScale),
			poop.Notes,
			poop.Timestamp.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write poop record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	// Upload to S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("text/csv"),
	})

	return err
}

func (s *S3Storage) GetAllPoopTags(ctx context.Context, userEmail string) ([]string, error) {
	// TODO: Implement S3 poop tags aggregation
	return []string{}, nil
}

func (s *S3Storage) GetAllFoodTags(ctx context.Context, userEmail string) ([]string, error) {
	// TODO: Implement S3 food tags aggregation
	return []string{"medicine"}, nil
}
