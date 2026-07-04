package database

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

//go:embed schema.sql
var schemaSQL string

//go:embed migrations/*.sql
var migrationsFS embed.FS

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

	// Create schema_migrations table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Automated Schema Creation
	var exists bool
	err = db.Get(&exists, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')")
	if err != nil {
		return fmt.Errorf("failed to check for tables: %w", err)
	}

	if !exists {
		fmt.Println("🚀 Database is empty. Running schema.sql...")
		_, err = db.Exec(schemaSQL)
		if err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
		fmt.Println("✅ Schema initialized successfully!")
	}

	// Apply migrations
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		log.Printf("No migrations found or error reading migrations dir: %v", err)
	} else {
		var migrationFiles []string
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
				migrationFiles = append(migrationFiles, entry.Name())
			}
		}
		sort.Strings(migrationFiles)

		for _, file := range migrationFiles {
			var applied bool
			err := db.Get(&applied, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", file)
			if err != nil {
				return fmt.Errorf("failed to check migration %s: %w", file, err)
			}

			if !applied {
				log.Printf("Applying migration: %s", file)
				content, err := migrationsFS.ReadFile(filepath.Join("migrations", file))
				if err != nil {
					return fmt.Errorf("failed to read migration %s: %w", file, err)
				}

				tx, err := db.Beginx()
				if err != nil {
					return err
				}

				if _, err := tx.Exec(string(content)); err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to apply migration %s: %w", file, err)
				}

				if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", file); err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to record migration %s: %w", file, err)
				}

				if err := tx.Commit(); err != nil {
					return fmt.Errorf("failed to commit migration %s: %w", file, err)
				}
				log.Printf("Successfully applied migration: %s", file)
			}
		}
	}

	return nil
}
