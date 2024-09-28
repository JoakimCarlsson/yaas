package repository

import (
	"context"
	"github.com/joakimcarlsson/yaas/internal/models"
)

type ActionRepository interface {
	CreateAction(ctx context.Context, action *models.Action) error
	GetActionByID(ctx context.Context, id int) (*models.Action, error)
	GetActionsByType(ctx context.Context, actionType string) ([]*models.Action, error)
	UpdateAction(ctx context.Context, action *models.Action) error
	DeleteAction(ctx context.Context, id int) error
	GetAllActions(ctx context.Context) ([]*models.Action, error)
}
