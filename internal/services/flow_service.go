package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/repository"
	"time"
)

type FlowService interface {
	InitiateFlow(ctx context.Context, flowType models.FlowType, requestURL string) (*models.Flow, error)
	GetFlowByID(ctx context.Context, id string) (*models.Flow, error)
	UpdateFlow(ctx context.Context, flow *models.Flow) error
	DeleteFlow(ctx context.Context, id string) error
	CleanupExpiredFlows(ctx context.Context) error
}

type flowService struct {
	flowRepo repository.FlowRepository
}

func NewFlowService(flowRepo repository.FlowRepository) FlowService {
	return &flowService{flowRepo: flowRepo}
}

func (s *flowService) InitiateFlow(ctx context.Context, flowType models.FlowType, requestURL string) (*models.Flow, error) {
	flowID := uuid.New().String()
	flow := &models.Flow{
		ID:         flowID,
		Type:       flowType,
		State:      models.FlowStateChooseMethod,
		IssuedAt:   time.Now(),
		ExpiresAt:  time.Now().Add(15 * time.Minute),
		RequestURL: requestURL,
	}
	err := s.flowRepo.CreateFlow(ctx, flow)
	return flow, err
}

func (s *flowService) GetFlowByID(ctx context.Context, id string) (*models.Flow, error) {
	return s.flowRepo.GetFlowByID(ctx, id)
}

func (s *flowService) UpdateFlow(ctx context.Context, flow *models.Flow) error {
	return s.flowRepo.UpdateFlow(ctx, flow)
}

func (s *flowService) DeleteFlow(ctx context.Context, id string) error {
	return s.flowRepo.DeleteFlow(ctx, id)
}

func (s *flowService) CleanupExpiredFlows(ctx context.Context) error {
	return s.flowRepo.CleanupExpiredFlows(ctx)
}
