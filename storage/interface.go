package storage

import (
	"context"

	"github.com/stinkyfingers/poopjournal/models"
)

type Storage interface {
	// Food operations
	SaveFood(ctx context.Context, food *models.Food) error
	ListFood(ctx context.Context, userEmail string) ([]*models.Food, error)
	UpdateFood(ctx context.Context, food *models.Food) error
	DeleteFood(ctx context.Context, userEmail, foodID string) error
	GetAllFoodTags(ctx context.Context, userEmail string) ([]string, error)

	// Poop operations
	SavePoop(ctx context.Context, poop *models.Poop) error
	ListPoop(ctx context.Context, userEmail string) ([]*models.Poop, error)
	UpdatePoop(ctx context.Context, poop *models.Poop) error
	DeletePoop(ctx context.Context, userEmail, poopID string) error
	GetAllPoopTags(ctx context.Context, userEmail string) ([]string, error)
}
