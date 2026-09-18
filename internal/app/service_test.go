package app

import (
	"context"
	"errors"
	"testing"

	"github.com/nikndip/projecthub-demo/internal/domain"
)

type repositoryStub struct {
	createdProject domain.Project
	createdRequest domain.PaymentRequest
	createCalls    int
	requestCalls   int
}

func (r *repositoryStub) CreateProject(_ context.Context, name, client string, budget int64) (domain.Project, error) {
	r.createCalls++
	r.createdProject = domain.Project{Name: name, Client: client, BudgetKopecks: budget}
	return r.createdProject, nil
}

func (r *repositoryStub) ListProjects(context.Context) ([]domain.Project, error) {
	return []domain.Project{r.createdProject}, nil
}

func (r *repositoryStub) CreatePaymentRequest(_ context.Context, projectID int64, purpose string, amount int64) (domain.PaymentRequest, error) {
	r.requestCalls++
	r.createdRequest = domain.PaymentRequest{ProjectID: projectID, Purpose: purpose, AmountKopecks: amount}
	return r.createdRequest, nil
}

func TestCreateProjectValidatesInput(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository)

	_, err := service.CreateProject(context.Background(), CreateProjectInput{Name: " ", Client: "Acme", BudgetKopecks: 10_000})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if repository.createCalls != 0 {
		t.Fatalf("repository called %d times for invalid input", repository.createCalls)
	}
}

func TestCreateProjectNormalizesInput(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository)

	project, err := service.CreateProject(context.Background(), CreateProjectInput{
		Name: "  Product launch  ", Client: "  Acme  ", BudgetKopecks: 1_500_000,
	})
	if err != nil {
		t.Fatalf("CreateProject returned error: %v", err)
	}
	if project.Name != "Product launch" || project.Client != "Acme" {
		t.Fatalf("input was not normalized: %#v", project)
	}
}

func TestCreatePaymentRequestValidatesAmount(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository)

	_, err := service.CreatePaymentRequest(context.Background(), 1, CreatePaymentRequestInput{Purpose: "Hosting", AmountKopecks: 0})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if repository.requestCalls != 0 {
		t.Fatalf("repository called %d times for invalid input", repository.requestCalls)
	}
}
