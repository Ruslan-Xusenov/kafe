package repository

import (
	"github.com/jmoiron/sqlx"
)

type SettingsRepository struct {
	db *sqlx.DB
}

func NewSettingsRepository(db *sqlx.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) Get(key string) (string, error) {
	var value string
	err := r.db.Get(&value, "SELECT value FROM settings WHERE key = $1", key)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *SettingsRepository) Set(key, value string) error {
	_, err := r.db.Exec("INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, NOW()) ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()", key, value)
	return err
}
