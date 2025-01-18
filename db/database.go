package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	Conn *sql.DB
}

func NewDatabase() (*Database, error) {
	// Ensure db directory exists
	dbDir := "./data"
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return nil, err
	}

	// Path to SQLite database file
	dbPath := filepath.Join(dbDir, "forum.db")

	// Open database connection
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable foreign key support
	_, err = conn.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	return &Database{Conn: conn}, nil
}

// Migrate runs SQL migration scripts
func (db *Database) Migrate() error {
	// Read and execute migration script
	migrationScript, err := os.ReadFile("./db/migrations/001_create_tables.sql")
	if err != nil {
		return err
	}

	// Execute migration script
	_, err = db.Conn.Exec(string(migrationScript))
	return err
}

// Close closes the database connection
func (db *Database) Close() error {
	return db.Conn.Close()
}
