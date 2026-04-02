package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/username/kafe-backend/internal/models"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `INSERT INTO users (full_name, phone, password_hash, role) 
              VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	
	err := r.db.QueryRow(query, user.FullName, user.Phone, user.PasswordHash, user.Role).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByPhone(phone string) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE phone = $1`
	err := r.db.Get(&user, query, phone)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get user by phone: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE id = $1`
	err := r.db.Get(&user, query, id)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	query := `UPDATE users SET full_name = $1, phone = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(query, user.FullName, user.Phone, user.ID)
	return err
}

func (r *UserRepository) GetStaff() ([]models.User, error) {
	var users []models.User
	query := `SELECT * FROM users WHERE role != 'customer' ORDER BY role ASC, full_name ASC`
	err := r.db.Select(&users, query)
	return users, err
}

func (r *UserRepository) Count() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users`
	err := r.db.Get(&count, query)
	return count, err
}
