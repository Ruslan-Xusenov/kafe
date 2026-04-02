package database

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func InitDB() error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db

	// Automated Schema Creation
	var exists bool
	err = db.Get(&exists, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')")
	if err != nil {
		return fmt.Errorf("failed to check for tables: %w", err)
	}

	if !exists {
		fmt.Println("🚀 Database is empty. Rooting schema.sql...")
		schema, err := os.ReadFile("internal/database/schema.sql")
		if err != nil {
			// Try relative to root if called from cmd/api
			schema, err = os.ReadFile("../../internal/database/schema.sql")
			if err != nil {
				return fmt.Errorf("failed to read schema file: %w", err)
			}
		}
		
		_, err = db.Exec(string(schema))
		if err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
		fmt.Println("✅ Schema initialized successfully!")
	}

	return nil
}
