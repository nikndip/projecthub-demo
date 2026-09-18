package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/nikndip/projecthub-demo/internal/domain"
)

type Repository interface {
	CreateProject(ctx context.Context, name, client string, budgetKopecks int64) (domain.Project, error)
	ListProjects(ctx context.Context) ([]domain.Project, error)
	CreatePaymentRequest(ctx context.Context, projectID int64, purpose string, amountKopecks int64) (domain.PaymentRequest, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

type CreateProjectInput struct {
	Name          string `json:"name"`
	Client        string `json:"client"`
	BudgetKopecks int64  `json:"budget_kopecks"`
}

func (s *Service) CreateProject(ctx context.Context, input CreateProjectInput) (domain.Project, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Client = strings.TrimSpace(input.Client)
	if input.Name == "" || input.Client == "" || input.BudgetKopecks <= 0 {
		return domain.Project{}, fmt.Errorf("%w: name, client and positive budget are required", domain.ErrInvalidInput)
	}
	return s.repository.CreateProject(ctx, input.Name, input.Client, input.BudgetKopecks)
}

func (s *Service) ListProjects(ctx context.Context) ([]domain.Project, error) {
	return s.repository.ListProjects(ctx)
}

type CreatePaymentRequestInput struct {
	Purpose       string `json:"purpose"`
	AmountKopecks int64  `json:"amount_kopecks"`
}

func (s *Service) CreatePaymentRequest(ctx context.Context, projectID int64, input CreatePaymentRequestInput) (domain.PaymentRequest, error) {
	input.Purpose = strings.TrimSpace(input.Purpose)
	if projectID <= 0 || input.Purpose == "" || input.AmountKopecks <= 0 {
		return domain.PaymentRequest{}, fmt.Errorf("%w: project, purpose and positive amount are required", domain.ErrInvalidInput)
	}
	return s.repository.CreatePaymentRequest(ctx, projectID, input.Purpose, input.AmountKopecks)
}
