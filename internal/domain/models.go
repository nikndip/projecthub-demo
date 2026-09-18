package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound       = errors.New("resource not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrBudgetExceeded = errors.New("project budget exceeded")
)

type Project struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Client          string    `json:"client"`
	BudgetKopecks   int64     `json:"budget_kopecks"`
	ReservedKopecks int64     `json:"reserved_kopecks"`
	CreatedAt       time.Time `json:"created_at"`
}

type PaymentRequest struct {
	ID            int64     `json:"id"`
	ProjectID     int64     `json:"project_id"`
	Purpose       string    `json:"purpose"`
	AmountKopecks int64     `json:"amount_kopecks"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
