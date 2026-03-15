package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/haydary1986/archiving-qa/internal/config"
)

type DB struct {
	*sql.DB
}

func New(cfg *config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connected successfully")
	return &DB{db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) RunMigrations() error {
	migrations := []string{
		migrationCreateExtensions,
		migrationCreateRolesTable,
		migrationCreatePermissionsTable,
		migrationCreateRolePermissionsTable,
		migrationCreateUsersTable,
		migrationCreateCategoriesTable,
		migrationCreateDocumentsTable,
		migrationCreateFilesTable,
		migrationCreateTagsTable,
		migrationCreateDocumentTagsTable,
		migrationCreatePersonsTable,
		migrationCreateDocumentPersonsTable,
		migrationCreateRoutingsTable,
		migrationCreateAuditLogTable,
		migrationCreateCustomFieldDefsTable,
		migrationCreateSystemSettingsTable,
		migrationCreateShareLinksTable,
		migrationCreateUserCategoryAccess,
		migrationCreateJobsTable,
		migrationSeedDefaultData,
	}

	for i, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}
