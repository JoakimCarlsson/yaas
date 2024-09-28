package services

import (
	"context"
	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/repository"
)

type ActionService interface {
	CreateAction(ctx context.Context, action *models.Action) error
	GetActionByID(ctx context.Context, id int) (*models.Action, error)
	GetActionsByType(ctx context.Context, actionType string) ([]*models.Action, error)
	UpdateAction(ctx context.Context, action *models.Action) error
	DeleteAction(ctx context.Context, id int) error
	GetAllActions(ctx context.Context) ([]*models.Action, error)
}

type actionService struct {
	actionRepo repository.ActionRepository
}

func NewActionService(actionRepo repository.ActionRepository) ActionService {
	return &actionService{
		actionRepo: actionRepo,
	}
}

func (s *actionService) CreateAction(ctx context.Context, action *models.Action) error {
	return s.actionRepo.CreateAction(ctx, action)
}

func (s *actionService) GetActionByID(ctx context.Context, id int) (*models.Action, error) {
	return s.actionRepo.GetActionByID(ctx, id)
}

func (s *actionService) GetActionsByType(ctx context.Context, actionType string) ([]*models.Action, error) {
	return s.actionRepo.GetActionsByType(ctx, actionType)
}

func (s *actionService) UpdateAction(ctx context.Context, action *models.Action) error {
	return s.actionRepo.UpdateAction(ctx, action)
}

func (s *actionService) DeleteAction(ctx context.Context, id int) error {
	return s.actionRepo.DeleteAction(ctx, id)
}

func (s *actionService) GetAllActions(ctx context.Context) ([]*models.Action, error) {
	return s.actionRepo.GetAllActions(ctx)
}
