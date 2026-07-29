package repository

import (
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/username/kafe-backend/internal/models"
)

type AuditRepository struct {
	db *sqlx.DB
}

func NewAuditRepository(db *sqlx.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(actorID *int, actorRole, action, entityType string, entityID *int, details interface{}, ipAddress string) error {
	if details == nil {
		details = map[string]interface{}{}
	}

	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		INSERT INTO audit_logs (actor_id, actor_role, action, entity_type, entity_id, details, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)
	`, actorID, actorRole, action, entityType, entityID, string(detailsJSON), ipAddress)
	return err
}

func (r *AuditRepository) List(limit int) ([]models.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	var logs []models.AuditLog
	err := r.db.Select(&logs, `
		SELECT 
			al.id,
			al.actor_id,
			COALESCE(u.full_name, '') as actor_name,
			COALESCE(al.actor_role, '') as actor_role,
			al.action,
			al.entity_type,
			al.entity_id,
			al.details::text as details,
			COALESCE(al.ip_address, '') as ip_address,
			al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON u.id = al.actor_id
		ORDER BY al.created_at DESC
		LIMIT $1
	`, limit)
	return logs, err
}
