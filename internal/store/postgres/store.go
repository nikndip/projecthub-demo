package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikndip/projecthub-demo/internal/domain"
)

//go:embed migrations/001_init.sql
var migrationSQL string

type Store struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, migrationSQL); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}
	return nil
}

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Store) CreateProject(ctx context.Context, name, client string, budgetKopecks int64) (domain.Project, error) {
	const query = `
		INSERT INTO projects (name, client, budget_kopecks)
		VALUES ($1, $2, $3)
		RETURNING id, name, client, budget_kopecks, reserved_kopecks, created_at`

	var project domain.Project
	err := s.pool.QueryRow(ctx, query, name, client, budgetKopecks).Scan(
		&project.ID, &project.Name, &project.Client, &project.BudgetKopecks,
		&project.ReservedKopecks, &project.CreatedAt,
	)
	if err != nil {
		return domain.Project{}, fmt.Errorf("create project: %w", err)
	}
	return project, nil
}

func (s *Store) ListProjects(ctx context.Context) ([]domain.Project, error) {
	const query = `
		SELECT id, name, client, budget_kopecks, reserved_kopecks, created_at
		FROM projects
		ORDER BY id DESC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	projects := make([]domain.Project, 0)
	for rows.Next() {
		var project domain.Project
		if err := rows.Scan(
			&project.ID, &project.Name, &project.Client, &project.BudgetKopecks,
			&project.ReservedKopecks, &project.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	return projects, nil
}

func (s *Store) CreatePaymentRequest(ctx context.Context, projectID int64, purpose string, amountKopecks int64) (domain.PaymentRequest, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return domain.PaymentRequest{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var budget, reserved int64
	err = tx.QueryRow(ctx, `
		SELECT budget_kopecks, reserved_kopecks
		FROM projects
		WHERE id = $1
		FOR UPDATE`, projectID).Scan(&budget, &reserved)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PaymentRequest{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.PaymentRequest{}, fmt.Errorf("lock project: %w", err)
	}
	if amountKopecks > budget-reserved {
		return domain.PaymentRequest{}, domain.ErrBudgetExceeded
	}

	const insertRequest = `
		INSERT INTO payment_requests (project_id, purpose, amount_kopecks)
		VALUES ($1, $2, $3)
		RETURNING id, project_id, purpose, amount_kopecks, status, created_at`
	var request domain.PaymentRequest
	if err := tx.QueryRow(ctx, insertRequest, projectID, purpose, amountKopecks).Scan(
		&request.ID, &request.ProjectID, &request.Purpose,
		&request.AmountKopecks, &request.Status, &request.CreatedAt,
	); err != nil {
		return domain.PaymentRequest{}, fmt.Errorf("insert payment request: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE projects
		SET reserved_kopecks = reserved_kopecks + $1
		WHERE id = $2`, amountKopecks, projectID); err != nil {
		return domain.PaymentRequest{}, fmt.Errorf("reserve project budget: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.PaymentRequest{}, fmt.Errorf("commit transaction: %w", err)
	}
	return request, nil
}
