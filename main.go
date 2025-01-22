package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"forum/db"
	"forum/handlers"
	"forum/middleware"
	"forum/repository"
	"forum/routes"

	"github.com/joho/godotenv"
)

func main() {
	// Detailed startup logging
	log.Println("🚀 Forum Application Startup Initiated")
	startupTimer := time.Now()

	// Load environment configuration with detailed logging
	if err := loadEnvironment(); err != nil {
		log.Fatalf("❌ Environment Configuration Failed: %v", err)
	}
	log.Printf("✅ Environment Configuration Loaded (Took %v)", time.Since(startupTimer))

	// Set up logging
	setupLogging()

	// Initialize database with comprehensive error tracking
	database, err := initializeDatabase()
	if err != nil {
		log.Fatalf("❌ Database Initialization Failed: %v\n%+v", err, err)
	}
	defer func() {
		if closeErr := database.Close(); closeErr != nil {
			log.Printf("⚠️ Database Close Error: %v", closeErr)
		}
	}()
	log.Printf("✅ Database Initialized (Took %v)", time.Since(startupTimer))

	// Run database migrations with detailed error context
	if err := runMigrations(database); err != nil {
		log.Fatalf("❌ Database Migration Failed: %v\n%+v", err, err)
	}
	log.Printf("✅ Database Migrations Completed (Took %v)", time.Since(startupTimer))

	// Initialize repositories
	userRepo := repository.NewUserRepository(database.Conn)
	postRepo := repository.NewPostRepository(database.Conn)
	sessionRepo := repository.NewSessionRepository(database.Conn)

	// Initialize auth middleware
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		log.Fatal("JWT_SECRET not set in environment")
	}
	authMiddleware := middleware.NewAuthMiddleware(sessionRepo, secretKey, 24)

	// Initialize template renderer
	templateRenderer := handlers.NewTemplateRenderer(authMiddleware)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, authMiddleware, templateRenderer)
	postHandler := handlers.NewPostHandler(postRepo, authMiddleware, templateRenderer)

	// Create router with comprehensive initialization
	router, err := createRouter(database, authHandler, postHandler, authMiddleware)
	if err != nil {
		log.Fatalf("❌ Router Initialization Failed: %v\n%+v", err, err)
	}
	log.Printf("✅ Router Configured (Took %v)", time.Since(startupTimer))

	// Get server configuration
	port := getServerPort()

	// Start server with detailed startup logging
	log.Printf("🚀 Attempting to start server on port %s", port)
	if err := router.Start(port); err != nil {
		log.Fatalf("❌ Server Startup Failed: %v\n%+v", err, err)
	}

	log.Printf("🎉 Application Startup Completed Successfully (Total Time: %v)", time.Since(startupTimer))
}

// Enhanced environment loading with multiple path support
func loadEnvironment() error {
	potentialPaths := []string{
		"./data/.env",
		"../data/.env",
		"/app/data/.env",
		os.Getenv("ENV_FILE_PATH"),
	}

	for _, path := range potentialPaths {
		if path == "" {
			continue
		}
		log.Printf("Attempting to load environment from: %s", path)
		if err := godotenv.Load(path); err == nil {
			log.Printf("✅ Environment loaded from %s", path)
			return nil
		}
	}

	return fmt.Errorf("no valid .env file found in potential locations")
}

// Comprehensive database initialization
func initializeDatabase() (*db.Database, error) {
	dbConnString := os.Getenv("DB_CONNECTION_STRING")
	if dbConnString == "" {
		dbConnString = "file:forum.db?_foreign_keys=on"
		log.Println("⚠️ Using default database connection string")
	}

	log.Printf("Initializing database with connection string: %s", dbConnString)
	database, err := db.NewDatabase(dbConnString)
	if err != nil {
		return nil, fmt.Errorf("database initialization error: %w", err)
	}

	return database, nil
}

// Enhanced migration process
func runMigrations(database *db.Database) error {
	migrationPath := os.Getenv("MIGRATION_PATH")
	if migrationPath == "" {
		migrationPath = "./db/migrations/001_create_tables.sql"
		log.Println("⚠️ Using default migration script path")
	}

	log.Printf("Running migrations from: %s", migrationPath)
	if err := database.Migrate(migrationPath); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

// Router creation with error handling
func createRouter(database *db.Database, authHandler *handlers.AuthHandler, postHandler *handlers.PostHandler, authMiddleware *middleware.AuthMiddleware) (*routes.Router, error) {
	log.Println("Creating application router")
	router := routes.NewRouter(database, authHandler, postHandler, authMiddleware)

	// Additional validation can be added here
	if router == nil {
		return nil, fmt.Errorf("router initialization returned nil")
	}

	return router, nil
}

// Server port configuration
func getServerPort() string {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
		log.Println("⚠️ Using default port 8080")
	}
	return ":" + port
}

// setupLogging configures application logging
func setupLogging() {
	logLevel := os.Getenv("LOG_LEVEL")
	logOutputPath := os.Getenv("LOG_OUTPUT_PATH")

	// Default to stdout
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Configure log level
	switch logLevel {
	case "debug":
		log.Println("Logging in debug mode")
	case "error":
		log.Println("Logging only errors")
	default:
		log.Println("Default logging level")
	}

	// Optional: Log to file if path specified
	if logOutputPath != "" {
		logFile, err := os.OpenFile(logOutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			log.Printf("Failed to open log file: %v", err)
		} else {
			log.SetOutput(logFile)
		}
	}
}
