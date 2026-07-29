package models

import "time"

type AuditLog struct {
	ID         int       `json:"id" db:"id"`
	ActorID    *int      `json:"actor_id" db:"actor_id"`
	ActorName  string    `json:"actor_name" db:"actor_name"`
	ActorRole  string    `json:"actor_role" db:"actor_role"`
	Action     string    `json:"action" db:"action"`
	EntityType string    `json:"entity_type" db:"entity_type"`
	EntityID   *int      `json:"entity_id" db:"entity_id"`
	Details    string    `json:"details" db:"details"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
