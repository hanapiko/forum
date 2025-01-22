package routes

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"forum/db"
	"forum/handlers"
	"forum/middleware"
	"forum/repository"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

// init loads environment variables before the application starts
func init() {
	// Load .env file from the data directory
	if err := godotenv.Load("data/.env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
}

type Router struct {
	r                  *chi.Mux
	authMiddleware     *middleware.AuthMiddleware
	postRepo           repository.PostRepositoryInterface
	userRepo           repository.UserRepository
	commentRepo        repository.CommentRepositoryInterface
	likeRepo           repository.LikeRepositoryInterface
	categoryRepo       repository.CategoryRepositoryInterface
	sessionRepo        repository.SessionRepositoryInterface
	interactionRepo    repository.InteractionRepositoryInterface
	postHandler        *handlers.PostHandler
	commentHandler     *handlers.CommentHandler
	interactionHandler *handlers.InteractionHandler
	templateRenderer   *handlers.TemplateRenderer
	authHandler        *handlers.AuthHandler
}

func NewRouter(database *db.Database, authHandler *handlers.AuthHandler, postHandler *handlers.PostHandler, authMiddleware *middleware.AuthMiddleware) *Router {
	log.Println(" Initializing Router")

	userRepo, err := initializeUserRepository(database)
	if err != nil {
		log.Fatalf("Failed to initialize user repository: %v", err)
	}

	postRepo, err := initializePostRepository(database)
	if err != nil {
		log.Fatalf("Failed to initialize post repository: %v", err)
	}

	categoryRepo, err := initializeCategoryRepository(database)
	if err != nil {
		log.Fatalf("Failed to initialize category repository: %v", err)
	}

	commentRepo, err := initializeCommentRepository(database)
	if err != nil {
		log.Fatalf(" Failed to initialize comment repository: %v", err)
	}

	likeRepo, err := initializeLikeRepository(database)
	if err != nil {
		log.Fatalf(" Failed to initialize like repository: %v", err)
	}

	sessionRepo, err := initializeSessionRepository(database)
	if err != nil {
		log.Fatalf(" Failed to initialize session repository: %v", err)
	}

	interactionRepo, err := initializeInteractionRepository(database)
	if err != nil {
		log.Fatalf("Failed to initialize interaction repository: %v", err)
	}

	templateRenderer := handlers.NewTemplateRenderer(authMiddleware)

	router := &Router{
		r:                chi.NewRouter(),
		authMiddleware:   authMiddleware,
		postRepo:         postRepo,
		userRepo:         userRepo,
		commentRepo:      commentRepo,
		likeRepo:         likeRepo,
		categoryRepo:     categoryRepo,
		sessionRepo:      sessionRepo,
		interactionRepo:  interactionRepo,
		postHandler:      postHandler,
		templateRenderer: templateRenderer,
		authHandler:      authHandler,
	}

	// Initialize handlers
	commentHandler := handlers.NewCommentHandler(commentRepo, authMiddleware)
	interactionHandler := handlers.NewInteractionHandler(
		interactionRepo,
		authMiddleware,
	)

	router.commentHandler = commentHandler
	router.interactionHandler = interactionHandler

	// Create Router struct
	router.registerRoutes()

	log.Println(" Router Initialized Successfully")
	return router
}

func initializeUserRepository(database *db.Database) (repository.UserRepository, error) {
	repo := repository.NewUserRepository(database.Conn)
	// Add any additional validation or initialization logic
	return repo, nil
}

func initializePostRepository(database *db.Database) (repository.PostRepositoryInterface, error) {
	repo := repository.NewPostRepository(database.Conn)
	// Add any additional validation or initialization logic
	return repo, nil
}

func initializeCategoryRepository(database *db.Database) (repository.CategoryRepositoryInterface, error) {
	repo := repository.NewCategoryRepository(database.Conn)
	// Add any additional validation or initialization logic
	return repo, nil
}

func initializeCommentRepository(database *db.Database) (repository.CommentRepositoryInterface, error) {
	repo := repository.NewCommentRepository(database.Conn)
	// Add any additional validation or initialization logic
	return repo, nil
}

func initializeLikeRepository(database *db.Database) (repository.LikeRepositoryInterface, error) {
	repo := repository.NewLikeRepository(database.Conn)
	// Add any additional validation or initialization logic
	return repo, nil
}

func initializeSessionRepository(database *db.Database) (repository.SessionRepositoryInterface, error) {
	repo := repository.NewSessionRepository(database.Conn)
	// Add any additional validation or initialization logic
	return repo, nil
}

func initializeInteractionRepository(database *db.Database) (repository.InteractionRepositoryInterface, error) {
	repo := repository.NewInteractionRepository(database.Conn)
	// Add any additional validation or initialization logic
	return repo, nil
}

func (router *Router) registerRoutes() {
	r := router.r

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(router.authMiddleware.ProtectRoute)

	// Add static file serving
	fileServer := http.FileServer(http.Dir("./frontend/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Authentication routes
	r.Get("/login", router.authHandler.LoginPage)
	r.Post("/login", router.authHandler.Login)
	r.Get("/register", router.authHandler.RegisterPage)
	r.Post("/register", router.authHandler.Register)
	r.Get("/logout", router.authHandler.Logout)

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(router.authMiddleware.ProtectRoute)

		// Dashboard
		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			data := handlers.TemplateData{
				PageTitle: "Dashboard",
			}
			router.templateRenderer.Render(w, r, "user/dashboard", data)
		})

		// Posts routes
		r.Get("/posts/create", router.postHandler.CreatePostPage)
		r.Post("/posts/create", router.postHandler.CreatePost)
	})

	// Public routes
	r.Get("/posts", router.postHandler.ListPosts)
	r.Get("/posts/{id}", router.postHandler.GetPost)

	// Error handling routes
	r.Get("/403", func(w http.ResponseWriter, r *http.Request) {
		router.templateRenderer.RenderError(w, r, http.StatusForbidden, nil)
	})
	r.Get("/500", func(w http.ResponseWriter, r *http.Request) {
		router.templateRenderer.RenderError(w, r, http.StatusInternalServerError, nil)
	})
}

func (router *Router) Start(addr string) error {
	// Ensure address starts with a colon
	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}

	log.Printf(" Starting server on %s", addr)
	log.Printf("Server accessible at http://localhost%s", addr)

	// Start the server
	if err := http.ListenAndServe(addr, router.r); err != nil {
		return fmt.Errorf("server startup failed: %w", err)
	}

	return nil
}
