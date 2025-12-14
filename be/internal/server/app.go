package server

import (
	"context"
	"log/slog"

	"github.com/escalopa/family-tree/internal/config"
	"github.com/escalopa/family-tree/internal/db"
	httpDelivery "github.com/escalopa/family-tree/internal/delivery/http"
	"github.com/escalopa/family-tree/internal/delivery/http/handler"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/pkg/oauth"
	"github.com/escalopa/family-tree/internal/pkg/s3"
	"github.com/escalopa/family-tree/internal/pkg/token"
	"github.com/escalopa/family-tree/internal/repository"
	"github.com/escalopa/family-tree/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// App holds all application dependencies
type App struct {
	cfg    *config.Config
	pool   *pgxpool.Pool
	engine *gin.Engine
}

// NewApp creates and initializes a new application
func NewApp(cfg *config.Config) (*App, error) {
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize database
	ctx := context.Background()
	pool, err := db.NewPool(ctx, &cfg.Database)
	if err != nil {
		return nil, err
	}

	slog.Info("App.NewApp: database connected")

	// Initialize S3 client
	s3Client, err := s3.NewS3Client(
		ctx,
		cfg.S3.Endpoint,
		cfg.S3.Region,
		cfg.S3.AccessKey,
		cfg.S3.SecretKey,
		cfg.S3.Bucket,
	)
	if err != nil {
		pool.Close()
		return nil, err
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
	router := httpDelivery.NewRouter(
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

	return &App{
		cfg:    cfg,
		pool:   pool,
		engine: engine,
	}, nil
}

// Close closes all application resources
func (a *App) Close() {
	if a.pool != nil {
		a.pool.Close()
	}
}
