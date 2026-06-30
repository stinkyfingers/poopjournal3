package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

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

func (s *S3Storage) GetUserData(ctx context.Context, userID string) (*models.UserData, error) {
	key := s.getUserKey(userID, "data.json")
	userData := models.NewUserData()
	// Try to get the file from S3
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			// File doesn't exist, return empty slice
			return &models.UserData{}, nil
		}
		return nil, fmt.Errorf("failed to get food file from S3: %w", err)
	}
	defer result.Body.Close()

	err = json.NewDecoder(result.Body).Decode(&userData)
	if err != nil {
		if err == io.EOF {
			return &models.UserData{}, nil
		}
		return nil, fmt.Errorf("failed to decode food JSON: %w", err)
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(userData.Foods, func(i, j int) bool {
		return userData.Foods[i].Timestamp.After(userData.Foods[j].Timestamp)
	})
	sort.Slice(userData.Poops, func(i, j int) bool {
		return userData.Poops[i].Timestamp.After(userData.Poops[j].Timestamp)
	})

	return userData, nil
}

func (s *S3Storage) SaveUserData(ctx context.Context, userID string, userData *models.UserData) error {
	key := s.getUserKey(userID, "data.json")

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(userData)
	if err != nil {
		return fmt.Errorf("failed to encode user data to JSON: %w", err)
	}

	// Upload to S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("application/json"),
	})

	return err
}
