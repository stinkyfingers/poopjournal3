package storage

import (
	"context"

	"github.com/stinkyfingers/poopjournal/models"
)

type Storage interface {
	GetUserData(ctx context.Context, userID string) (*models.UserData, error)
	SaveUserData(ctx context.Context, userID string, userData *models.UserData) error
}
