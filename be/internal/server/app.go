package server

import (
	"context"
	"log/slog"

	"github.com/escalopa/family-tree/internal/config"
	"github.com/escalopa/family-tree/internal/db"
	"github.com/escalopa/family-tree/internal/delivery/http"
	"github.com/escalopa/family-tree/internal/delivery/http/cookie"
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

type App struct {
	cfg    *config.Config
	pool   *pgxpool.Pool
	engine *gin.Engine
}

func NewApp(cfg *config.Config) (*App, error) {
	gin.SetMode(cfg.Server.Mode)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, &cfg.Database)
	if err != nil {
		return nil, err
	}

	slog.Info("App.NewApp: database connected")

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

	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	oauthStateRepo := repository.NewOAuthStateRepository(pool)
	memberRepo := repository.NewMemberRepository(pool)
	spouseRepo := repository.NewSpouseRepository(pool)
	historyRepo := repository.NewHistoryRepository(pool)
	scoreRepo := repository.NewScoreRepository(pool)
	roleRepo := repository.NewRoleRepository(pool)
	_ = roleRepo // May be used later

	oauthManager := oauth.NewOAuthManager(&cfg.OAuth)

	tokenMgr := token.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)

	cookieManager := cookie.NewManager(&cfg.Server.Cookie)

	authUseCase := usecase.NewAuthUseCase(userRepo, sessionRepo, oauthStateRepo, oauthManager, tokenMgr)
	userUseCase := usecase.NewUserUseCase(userRepo, scoreRepo, historyRepo)
	memberUseCase := usecase.NewMemberUseCase(memberRepo, spouseRepo, historyRepo, scoreRepo, s3Client)
	spouseUseCase := usecase.NewSpouseUseCase(spouseRepo, memberRepo, historyRepo, scoreRepo)
	treeUseCase := usecase.NewTreeUseCase(memberRepo, spouseRepo)

	authHandler := handler.NewAuthHandler(authUseCase, cookieManager)
	userHandler := handler.NewUserHandler(userUseCase)
	memberHandler := handler.NewMemberHandler(memberUseCase)
	spouseHandler := handler.NewSpouseHandler(spouseUseCase)
	treeHandler := handler.NewTreeHandler(treeUseCase)

	authMiddleware := middleware.NewAuthMiddleware(tokenMgr, authUseCase, cookieManager)

	router := http.NewRouter(
		authHandler,
		userHandler,
		memberHandler,
		spouseHandler,
		treeHandler,
		authMiddleware,
		cfg.Server.AllowedOrigins,
	)

	engine := gin.New()
	engine.Use(gin.Recovery())
	router.Setup(engine)

	return &App{
		cfg:    cfg,
		pool:   pool,
		engine: engine,
	}, nil
}

func (a *App) Close() {
	if a.pool != nil {
		a.pool.Close()
	}
}
