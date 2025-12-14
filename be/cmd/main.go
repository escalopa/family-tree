package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"

	"github.com/escalopa/family-tree/internal/config"
	"github.com/escalopa/family-tree/internal/db"
	"github.com/escalopa/family-tree/internal/delivery/http"
	"github.com/escalopa/family-tree/internal/delivery/http/handler"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/pkg/oauth"
	"github.com/escalopa/family-tree/internal/pkg/s3"
	"github.com/escalopa/family-tree/internal/pkg/token"
	"github.com/escalopa/family-tree/internal/repository"
	"github.com/escalopa/family-tree/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize database
	ctx := context.Background()
	pool, err := db.NewPool(ctx, &cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Database connected successfully")

	// Initialize S3 client
	s3Client, err := s3.NewS3Client(
		cfg.S3.Endpoint,
		cfg.S3.Region,
		cfg.S3.AccessKey,
		cfg.S3.SecretKey,
		cfg.S3.Bucket,
	)
	if err != nil {
		log.Fatalf("Failed to initialize S3 client: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	memberRepo := repository.NewMemberRepository(pool)
	spouseRepo := repository.NewSpouseRepository(pool)
	historyRepo := repository.NewHistoryRepository(pool)
	scoreRepo := repository.NewScoreRepository(pool)
	roleRepo := repository.NewRoleRepository(pool)
	_ = roleRepo // May be used later

	// Initialize OAuth
	oauthManager := oauth.NewOAuthManager(&cfg.OAuth)

	// Initialize token manager
	tokenMgr := token.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(userRepo, sessionRepo, oauthManager, tokenMgr)
	userUseCase := usecase.NewUserUseCase(userRepo, scoreRepo, historyRepo)
	memberUseCase := usecase.NewMemberUseCase(memberRepo, spouseRepo, historyRepo, scoreRepo, s3Client)
	spouseUseCase := usecase.NewSpouseUseCase(spouseRepo, memberRepo, historyRepo, scoreRepo)
	treeUseCase := usecase.NewTreeUseCase(memberRepo, spouseRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUseCase)
	userHandler := handler.NewUserHandler(userUseCase)
	memberHandler := handler.NewMemberHandler(memberUseCase)
	spouseHandler := handler.NewSpouseHandler(spouseUseCase)
	treeHandler := handler.NewTreeHandler(treeUseCase)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(tokenMgr, authUseCase)

	// Initialize router
	router := http.NewRouter(
		authHandler,
		userHandler,
		memberHandler,
		spouseHandler,
		treeHandler,
		authMiddleware,
	)

	// Setup Gin engine
	engine := gin.Default()
	router.Setup(engine)

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.Server.Port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
