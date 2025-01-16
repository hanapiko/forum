package routes

import (
	"log"
	"net/http"

	"forum/db"
	"forum/handlers"
	"forum/middleware"
	"forum/repository"
)

type Router struct {
	mux             *http.ServeMux
	userRepo        *repository.UserRepository
	authMiddleware  *middleware.AuthMiddleware
	authHandler     *handlers.AuthHandler
}

func NewRouter(database *db.Database) *Router {
	// Initialize repositories
	userRepo := repository.NewUserRepository(database.Conn)

	// Secret key for JWT (in production, use a secure, environment-based method)
	secretKey := "your-secret-key-here"

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(userRepo, secretKey)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, authMiddleware)

	// Create router
	mux := http.NewServeMux()

	// Authentication Routes
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/logout", authMiddleware.ProtectRoute(authHandler.Logout))

	return &Router{
		mux:             mux,
		userRepo:        userRepo,
		authMiddleware:  authMiddleware,
		authHandler:     authHandler,
	}
}

func (r *Router) Start(port string) {
	log.Printf("🚀 Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r.mux))
}