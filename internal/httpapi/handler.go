package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/nikndip/projecthub-demo/internal/app"
	"github.com/nikndip/projecthub-demo/internal/domain"
)

type Service interface {
	CreateProject(ctx context.Context, input app.CreateProjectInput) (domain.Project, error)
	ListProjects(ctx context.Context) ([]domain.Project, error)
	CreatePaymentRequest(ctx context.Context, projectID int64, input app.CreatePaymentRequestInput) (domain.PaymentRequest, error)
}

type HealthChecker interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	service Service
	health  HealthChecker
	logger  *slog.Logger
}

func NewHandler(service Service, health HealthChecker, logger *slog.Logger) *Handler {
	return &Handler{service: service, health: health, logger: logger}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("GET /api/v1/projects", h.listProjects)
	mux.HandleFunc("POST /api/v1/projects", h.createProject)
	mux.HandleFunc("POST /api/v1/projects/{projectID}/payment-requests", h.createPaymentRequest)
	return h.logging(mux)
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	if err := h.health.Ping(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.ListProjects(r.Context())
	if err != nil {
		h.logger.Error("list projects", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": projects})
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var input app.CreateProjectInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	project, err := h.service.CreateProject(r.Context(), input)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (h *Handler) createPaymentRequest(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.ParseInt(r.PathValue("projectID"), 10, 64)
	if err != nil || projectID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	var input app.CreatePaymentRequestInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	request, err := h.service.CreatePaymentRequest(r.Context(), projectID, input)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, request)
}

func (h *Handler) writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrBudgetExceeded):
		writeError(w, http.StatusConflict, err.Error())
	default:
		h.logger.Error("request failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *Handler) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.logger.Info("http request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func decodeJSON(r *http.Request, destination any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(destination)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
