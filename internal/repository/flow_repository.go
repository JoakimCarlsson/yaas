package repository

import (
	"context"
	"github.com/joakimcarlsson/yaas/internal/models"
)

type FlowRepository interface {
	CreateFlow(ctx context.Context, flow *models.Flow) error
	GetFlowByID(ctx context.Context, id string) (*models.Flow, error)
	UpdateFlow(ctx context.Context, flow *models.Flow) error
	DeleteFlow(ctx context.Context, id string) error
	CleanupExpiredFlows(ctx context.Context) error
}
