package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nikndip/projecthub-demo/internal/app"
	"github.com/nikndip/projecthub-demo/internal/domain"
)

type serviceStub struct {
	projects []domain.Project
}

func (s *serviceStub) CreateProject(_ context.Context, input app.CreateProjectInput) (domain.Project, error) {
	return domain.Project{ID: 1, Name: input.Name, Client: input.Client, BudgetKopecks: input.BudgetKopecks}, nil
}

func (s *serviceStub) ListProjects(context.Context) ([]domain.Project, error) {
	return s.projects, nil
}

func (s *serviceStub) CreatePaymentRequest(_ context.Context, projectID int64, input app.CreatePaymentRequestInput) (domain.PaymentRequest, error) {
	return domain.PaymentRequest{ID: 1, ProjectID: projectID, Purpose: input.Purpose, AmountKopecks: input.AmountKopecks}, nil
}

type healthStub struct{}

func (healthStub) Ping(context.Context) error { return nil }

func TestCreateProject(t *testing.T) {
	handler := NewHandler(&serviceStub{}, healthStub{}, slog.New(slog.NewTextHandler(io.Discard, nil))).Routes()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(`{"name":"Launch","client":"Acme","budget_kopecks":500000}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"name":"Launch"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestRejectsUnknownJSONField(t *testing.T) {
	handler := NewHandler(&serviceStub{}, healthStub{}, slog.New(slog.NewTextHandler(io.Discard, nil))).Routes()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(`{"name":"Launch","client":"Acme","budget_kopecks":500000,"unknown":true}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
