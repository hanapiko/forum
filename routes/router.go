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
	postRepo        *repository.PostRepository
	categoryRepo    *repository.CategoryRepository
	authMiddleware  *middleware.AuthMiddleware
	authHandler     *handlers.AuthHandler
	postHandler     *handlers.PostHandler
	categoryHandler *handlers.CategoryHandler
}

func NewRouter(database *db.Database) *Router {
	// Initialize repositories
	userRepo := repository.NewUserRepository(database.Conn)
	postRepo := repository.NewPostRepository(database.Conn)
	categoryRepo := repository.NewCategoryRepository(database.Conn)
	sessionRepo := repository.NewSessionRepository(database.Conn)

	// Secret key for JWT (in production, use a secure, environment-based method)
	secretKey := "your-secret-key-here"

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(sessionRepo, secretKey)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, authMiddleware)
	postHandler := handlers.NewPostHandler(postRepo, authMiddleware)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)

	// Create router
	mux := http.NewServeMux()

	// Authentication Routes
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/logout", authMiddleware.ProtectRoute(authHandler.Logout))

	// Post Routes
	mux.HandleFunc("/posts/create", authMiddleware.ProtectRoute(postHandler.CreatePost))
	mux.HandleFunc("/posts/list", postHandler.ListPosts)
	mux.HandleFunc("/posts/get", postHandler.GetPost)
	mux.HandleFunc("/posts/update", authMiddleware.ProtectRoute(postHandler.UpdatePost))
	mux.HandleFunc("/posts/delete", authMiddleware.ProtectRoute(postHandler.DeletePost))

	// Category Routes
	mux.HandleFunc("/categories/create", authMiddleware.ProtectRoute(categoryHandler.CreateCategory))
	mux.HandleFunc("/categories/list", categoryHandler.ListCategories)
	mux.HandleFunc("/categories/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			categoryHandler.GetCategory(w, r)
		case http.MethodPut:
			categoryHandler.UpdateCategory(w, r)
		case http.MethodDelete:
			categoryHandler.DeleteCategory(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return &Router{
		mux:             mux,
		userRepo:        userRepo,
		postRepo:        postRepo,
		categoryRepo:    categoryRepo,
		authMiddleware:  authMiddleware,
		authHandler:     authHandler,
		postHandler:     postHandler,
		categoryHandler: categoryHandler,
	}
}

func (r *Router) Start(port string) {
	log.Printf("🚀 Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r.mux))
}
