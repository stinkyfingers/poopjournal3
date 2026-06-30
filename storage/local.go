package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

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

func (l *LocalStorage) getUserFileName(userID string) (string, error) {
	filename := filepath.Join(l.getUserDir(userID), "data.json")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		_, err := os.Create(filename)
		if err != nil {
			return "", fmt.Errorf("failed to create user data file: %w", err)
		}
		return filename, nil
	}
	return filename, nil
}

func (l *LocalStorage) GetUserData(ctx context.Context, userID string) (*models.UserData, error) {
	fn, err := l.getUserFileName(userID)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	userData := models.NewUserData()
	err = json.NewDecoder(f).Decode(userData)
	if err != nil {
		if err == io.EOF {
			return userData, nil
		}
		return nil, err
	}

	return userData, err
}

func (l *LocalStorage) SaveUserData(ctx context.Context, userID string, userData *models.UserData) error {
	fn, err := l.getUserFileName(userID)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(userData)
}
