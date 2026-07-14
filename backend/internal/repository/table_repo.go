package repository

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/username/kafe-backend/internal/models"
)

type TableRepository struct {
	db *sqlx.DB
}

func NewTableRepository(db *sqlx.DB) *TableRepository {
	return &TableRepository{db: db}
}

func (r *TableRepository) GetAll() ([]models.Table, error) {
	var tables []models.Table
	err := r.db.Select(&tables, `SELECT * FROM tables ORDER BY NULLIF(regexp_replace(name, '\D', '', 'g'), '')::int NULLS LAST, name ASC`)
	if err != nil {
		return nil, err
	}
	return tables, nil
}

func (r *TableRepository) GetByID(id int) (*models.Table, error) {
	var table models.Table
	err := r.db.Get(&table, "SELECT * FROM tables WHERE id = $1", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &table, err
}

func (r *TableRepository) Create(table *models.Table) error {
	query := `INSERT INTO tables (name, capacity, status) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.db.QueryRow(query, table.Name, table.Capacity, table.Status).Scan(&table.ID, &table.CreatedAt)
}

func (r *TableRepository) Update(table *models.Table) error {
	query := `UPDATE tables SET name = $1, capacity = $2, status = $3 WHERE id = $4`
	_, err := r.db.Exec(query, table.Name, table.Capacity, table.Status, table.ID)
	return err
}

func (r *TableRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM tables WHERE id = $1", id)
	return err
}
