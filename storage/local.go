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

func (l *LocalStorage) getFoods(userID string) ([]*models.Food, error) {
	foodFile := filepath.Join(l.getUserDir(userID), "food.json")

	if _, err := os.Stat(foodFile); os.IsNotExist(err) {
		return []*models.Food{}, nil
	}

	f, err := os.Open(foodFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open food file: %w", err)
	}
	defer f.Close()
	var userFoods []*models.Food
	err = json.NewDecoder(f).Decode(&userFoods)
	if err != nil {
		if err == io.EOF {
			return []*models.Food{}, nil
		}
		return nil, err
	}
	return userFoods, nil
}
func (l *LocalStorage) getFoodFile(userID string) (*os.File, error) {
	foodFile := filepath.Join(l.getUserDir(userID), "food.json")
	if _, err := os.Stat(foodFile); os.IsNotExist(err) {
		f, err := os.Create(foodFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create food file: %w", err)
		}
		return f, nil
	}
	return os.OpenFile(foodFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}

func (l *LocalStorage) getPoops(userID string) ([]*models.Poop, error) {
	poopFile := filepath.Join(l.getUserDir(userID), "poop.json")

	if _, err := os.Stat(poopFile); os.IsNotExist(err) {
		return []*models.Poop{}, nil
	}

	f, err := os.Open(poopFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open poop file: %w", err)
	}
	defer f.Close()
	var userPoops []*models.Poop
	err = json.NewDecoder(f).Decode(&userPoops)
	if err != nil {
		if err == io.EOF {
			return []*models.Poop{}, nil
		}
		return nil, err
	}
	return userPoops, nil
}

func (l *LocalStorage) getPoopFile(userID string) (*os.File, error) {
	poopFile := filepath.Join(l.getUserDir(userID), "poop.json")
	if _, err := os.Stat(poopFile); os.IsNotExist(err) {
		f, err := os.Create(poopFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create poop file: %w", err)
		}
		return f, nil
	}
	return os.OpenFile(poopFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}

func (l *LocalStorage) SaveFood(ctx context.Context, food *models.Food) error {
	userFoods, err := l.getFoods(food.UserID)
	if err != nil {
		return err
	}

	f, err := l.getFoodFile(food.UserID)
	if err != nil {
		return err
	}
	defer f.Close()
	userFoods = append(userFoods, food)
	return json.NewEncoder(f).Encode(userFoods)
}

func (l *LocalStorage) ListFood(ctx context.Context, userID string) ([]*models.Food, error) {
	return l.getFoods(userID)
}

func (l *LocalStorage) UpdateFood(ctx context.Context, food *models.Food) error {
	userFoods, err := l.getFoods(food.UserID)
	if err != nil {
		return err
	}

	f, err := l.getFoodFile(food.UserID)
	if err != nil {
		return err
	}
	defer f.Close()
	for i := range userFoods {
		if userFoods[i].ID == food.ID {
			userFoods[i] = food
		}
	}
	return json.NewEncoder(f).Encode(userFoods)
}

func (l *LocalStorage) DeleteFood(ctx context.Context, userID, foodID string) error {
	userFoods, err := l.getFoods(userID)
	if err != nil {
		return err
	}

	f, err := l.getFoodFile(userID)
	if err != nil {
		return err
	}
	defer f.Close()
	filteredFoods := make([]*models.Food, 0, len(userFoods))
	for i := range userFoods {
		if userFoods[i].ID != foodID {
			filteredFoods = append(filteredFoods, userFoods[i])
		}
	}
	return json.NewEncoder(f).Encode(filteredFoods)
}

func (l *LocalStorage) GetAllFoodTags(ctx context.Context, userID string) ([]string, error) {
	userFoods, err := l.getFoods(userID)
	if err != nil {
		return nil, err
	}

	var tags []string
	for _, food := range userFoods {
		tags = append(tags, food.Tags...)
	}
	return tags, nil
}

func (l *LocalStorage) SavePoop(ctx context.Context, poop *models.Poop) error {
	userPoops, err := l.getPoops(poop.UserID)
	if err != nil {
		return err
	}

	f, err := l.getPoopFile(poop.UserID)
	if err != nil {
		return err
	}
	defer f.Close()
	userPoops = append(userPoops, poop)
	return json.NewEncoder(f).Encode(userPoops)
}

func (l *LocalStorage) ListPoop(ctx context.Context, userID string) ([]*models.Poop, error) {
	return l.getPoops(userID)
}
func (l *LocalStorage) UpdatePoop(ctx context.Context, poop *models.Poop) error {
	userPoops, err := l.getPoops(poop.UserID)
	if err != nil {
		return err
	}

	f, err := l.getPoopFile(poop.UserID)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&userPoops)
	if err != nil {
		return err
	}
	for i := range userPoops {
		if userPoops[i].ID == poop.ID {
			userPoops[i] = poop
		}
	}
	return json.NewEncoder(f).Encode(userPoops)
}

func (l *LocalStorage) DeletePoop(ctx context.Context, userID, poopID string) error {
	userPoops, err := l.getPoops(userID)
	if err != nil {
		return err
	}

	f, err := l.getPoopFile(userID)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&userPoops)
	if err != nil {
		return err
	}
	filteredPoops := make([]*models.Poop, 0, len(userPoops))
	for i := range userPoops {
		if userPoops[i].ID != poopID {
			filteredPoops = append(filteredPoops, userPoops[i])
		}
	}
	return json.NewEncoder(f).Encode(filteredPoops)
}

func (l *LocalStorage) GetAllPoopTags(ctx context.Context, userID string) ([]string, error) {
	userPoops, err := l.getPoops(userID)
	if err != nil {
		return nil, err
	}
	var tags []string
	for _, food := range userPoops {
		tags = append(tags, food.Tags...)
	}
	return tags, nil
}
