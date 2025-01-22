package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	Conn *sql.DB
}

func NewDatabase(connString string) (*Database, error) {
	// Validate connection string
	if connString == "" {
		return nil, fmt.Errorf("database connection string cannot be empty")
	}

	log.Printf("Attempting to open database connection: %s", connString)

	// Open database connection
	conn, err := sql.Open("sqlite3", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Test the connection
	if err = conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Enable foreign key support
	_, err = conn.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %v", err)
	}

	log.Printf("✅ Database connection established successfully: %s", connString)
	return &Database{Conn: conn}, nil
}

// Migrate runs SQL migration scripts from a specified path
func (db *Database) Migrate(migrationPath string) error {
	// Validate migration path
	if migrationPath == "" {
		return fmt.Errorf("migration script path cannot be empty")
	}

	log.Printf("Attempting to run migrations from: %s", migrationPath)

	// Read migration script
	migrationScript, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration script: %w", err)
	}

	// Split script into individual statements
	statements := strings.Split(string(migrationScript), ";")

	// Begin a transaction for migration
	tx, err := db.Conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback in case of error

	// Execute each migration statement
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		log.Printf("Executing migration statement: %s", statement)
		_, err = tx.Exec(statement)
		if err != nil {
			return fmt.Errorf("migration statement execution failed: %w\nStatement: %s", err, statement)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	log.Printf("✅ Database migration completed successfully from %s", migrationPath)
	return nil
}

// Close safely closes the database connection
func (db *Database) Close() error {
	if db.Conn != nil {
		log.Println("Closing database connection")
		return db.Conn.Close()
	}
	return nil
}
