package repository

import (
	"context"
	"encoding/json"
	"errors"

	"task-decomposer-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Repository { return &Repository{db: db} }
func (r *Repository) Save(ctx context.Context, requestID string, req domain.PlanRequest, response domain.PlanResponse, rawHash string) error {
	if response.Plan == nil {
		return nil
	}
	planID := response.Plan.ID
	if planID == "" {
		planID = uuid.NewString()
		response.Plan.ID = planID
	}
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO plans(id, request_id, summary, current_version) VALUES($1,$2,$3,$4) ON CONFLICT(id) DO NOTHING`,
		planID, requestID, response.Plan.Summary, response.Plan.Version)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO plan_versions(plan_id, version, response_json, context_hash) VALUES($1,$2,$3,$4) ON CONFLICT(plan_id,version) DO NOTHING`,
		planID, response.Plan.Version, payload, rawHash)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO plan_requests(request_id,payload_hash,status,response_json) VALUES($1,$2,'completed',$3)
		ON CONFLICT(request_id) DO UPDATE SET status='completed', response_json=EXCLUDED.response_json`, requestID, rawHash, payload)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r *Repository) Get(ctx context.Context, id string) ([]byte, error) {
	var b []byte
	err := r.db.QueryRow(ctx, `SELECT response_json FROM plan_versions WHERE plan_id=$1 ORDER BY version DESC LIMIT 1`, id).Scan(&b)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	return b, err
}
