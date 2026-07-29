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
	err := r.db.Select(&tables, `
		SELECT t.*, 
		       (SELECT waiter_id FROM orders WHERE table_id = t.id AND status IN ('new', 'preparing', 'ready') LIMIT 1) as active_waiter_id
		FROM tables t
		ORDER BY 
			(t.status = 'free') DESC,
			(t.name !~ '^\s*\d+\s*$'), 
			NULLIF(regexp_replace(t.name, '\D', '', 'g'), '')::int NULLS LAST, 
			t.name ASC
	`)
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
	query := `INSERT INTO tables (name, capacity, status, pos_x, pos_y, shape, width, height, floor, rotation) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, created_at`
	return r.db.QueryRow(query, table.Name, table.Capacity, table.Status,
		table.PosX, table.PosY, table.Shape, table.Width, table.Height, table.Floor, table.Rotation,
	).Scan(&table.ID, &table.CreatedAt)
}

func (r *TableRepository) Update(table *models.Table) error {
	query := `UPDATE tables SET name = $1, capacity = $2, status = $3, 
		pos_x = $4, pos_y = $5, shape = $6, width = $7, height = $8, floor = $9, rotation = $10
		WHERE id = $11`
	_, err := r.db.Exec(query, table.Name, table.Capacity, table.Status,
		table.PosX, table.PosY, table.Shape, table.Width, table.Height, table.Floor, table.Rotation,
		table.ID)
	return err
}

func (r *TableRepository) Delete(id int) error {
	query := `DELETE FROM tables WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *TableRepository) UpdateStatus(id int, status string) error {
	query := `UPDATE tables SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

// UpdateLayout batch-updates table positions (for drag-and-drop floor plan editor)
func (r *TableRepository) UpdateLayout(layouts []models.Table) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	for _, t := range layouts {
		_, err := tx.Exec(
			`UPDATE tables SET pos_x = $1, pos_y = $2, shape = $3, width = $4, height = $5, floor = $6, rotation = $7 WHERE id = $8`,
			t.PosX, t.PosY, t.Shape, t.Width, t.Height, t.Floor, t.Rotation, t.ID,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// GetByFloor retrieves tables for a specific floor
func (r *TableRepository) GetByFloor(floor int) ([]models.Table, error) {
	var tables []models.Table
	err := r.db.Select(&tables, `
		SELECT t.*, 
		       (SELECT waiter_id FROM orders WHERE table_id = t.id AND status IN ('new', 'preparing', 'ready') LIMIT 1) as active_waiter_id
		FROM tables t
		WHERE t.floor = $1
		ORDER BY t.name ASC
	`, floor)
	if err != nil {
		return nil, err
	}
	return tables, nil
}
