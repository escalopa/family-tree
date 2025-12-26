package server

import (
	"context"
	"log/slog"
	"slices"

	"github.com/escalopa/family-tree/internal/config"
	"github.com/escalopa/family-tree/internal/db"
	"github.com/escalopa/family-tree/internal/delivery/http"
	"github.com/escalopa/family-tree/internal/delivery/http/cookie"
	"github.com/escalopa/family-tree/internal/delivery/http/handler"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/pkg/oauth"
	"github.com/escalopa/family-tree/internal/pkg/ratelimit"
	"github.com/escalopa/family-tree/internal/pkg/redis"
	"github.com/escalopa/family-tree/internal/pkg/s3"
	"github.com/escalopa/family-tree/internal/pkg/token"
	"github.com/escalopa/family-tree/internal/repository"
	"github.com/escalopa/family-tree/internal/usecase"
	"github.com/escalopa/family-tree/internal/usecase/validator"
	"github.com/escalopa/family-tree/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	cfg          *config.Config
	pool         *pgxpool.Pool
	engine       *gin.Engine
	cleanupFuncs []func()
}

func NewApp(cfg *config.Config) (*App, error) {
	gin.SetMode(cfg.Server.Mode)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, &cfg.Database)
	if err != nil {
		return nil, err
	}

	slog.Info("App.NewApp: database connected")

	redisClient, err := redis.NewClient(ctx, cfg.Redis.URI)
	if err != nil {
		return nil, err
	}

	slog.Info("App.NewApp: redis connected")

	s3Client, err := s3.NewS3Client(
		ctx,
		cfg.S3.Endpoint,
		cfg.S3.Region,
		cfg.S3.AccessKey,
		cfg.S3.SecretKey,
		cfg.S3.Bucket,
		cfg.Upload.MaxImageSize,
		cfg.Upload.AllowedImageExts,
	)
	if err != nil {
		return nil, err
	}

	langRepo := repository.NewLanguageRepository(pool)
	if err := langRepo.InitializeLanguages(ctx); err != nil {
		return nil, err
	}

	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	oauthStateRepo := repository.NewOAuthStateRepository(pool)
	langPrefRepo := repository.NewUserLanguagePreferenceRepository(pool)
	memberRepo := repository.NewMemberRepository(pool)
	spouseRepo := repository.NewSpouseRepository(pool)
	historyRepo := repository.NewHistoryRepository(pool)
	scoreRepo := repository.NewScoreRepository(pool)
	roleRepo := repository.NewRoleRepository(pool)
	_ = roleRepo // May be used later

	txManager := repository.NewTransactionManager(pool)

	oauthManager := oauth.NewOAuthManager(&cfg.OAuth)

	tokenMgr := token.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)

	cookieManager := cookie.NewManager(&cfg.Server.Cookie)

	// Create validators
	marriageValidator := validator.NewMarriageValidator(memberRepo, spouseRepo)
	birthDateValidator := validator.NewBirthDateValidator(memberRepo, spouseRepo)
	relationshipValidator := validator.NewRelationshipValidator(memberRepo, spouseRepo)

	authUseCase := usecase.NewAuthUseCase(userRepo, sessionRepo, oauthStateRepo, oauthManager, tokenMgr)
	userUseCase := usecase.NewUserUseCase(userRepo, scoreRepo, historyRepo)
	memberUseCase := usecase.NewMemberUseCase(memberRepo, spouseRepo, historyRepo, scoreRepo, s3Client, txManager, marriageValidator, birthDateValidator, relationshipValidator)
	spouseUseCase := usecase.NewSpouseUseCase(spouseRepo, memberRepo, historyRepo, scoreRepo, txManager, marriageValidator)
	treeUseCase := usecase.NewTreeUseCase(memberRepo, spouseRepo)
	languageUseCase := usecase.NewLanguageUseCase(langRepo, langPrefRepo)

	authHandler := handler.NewAuthHandler(authUseCase, userUseCase, cookieManager)
	userHandler := handler.NewUserHandler(userUseCase)
	memberHandler := handler.NewMemberHandler(memberUseCase, languageUseCase)
	spouseHandler := handler.NewSpouseHandler(spouseUseCase)
	treeHandler := handler.NewTreeHandler(treeUseCase)
	languageHandler := handler.NewLanguageHandler(languageUseCase)

	authMiddleware := middleware.NewAuthMiddleware(tokenMgr, authUseCase, userRepo, cookieManager)

	authLimiterMiddleware := middleware.NewRateLimiter(
		ratelimit.New(redisClient, cfg.RateLimit.Auth),
		cfg.RateLimit.Auth.Enabled,
	)
	apiLimiterMiddleware := middleware.NewRateLimiter(
		ratelimit.New(redisClient, cfg.RateLimit.API),
		cfg.RateLimit.API.Enabled,
	)
	uploadLimiterMiddleware := middleware.NewRateLimiter(
		ratelimit.New(redisClient, cfg.RateLimit.Upload),
		cfg.RateLimit.Upload.Enabled,
	)

	router := http.NewRouter(
		authHandler,
		userHandler,
		memberHandler,
		spouseHandler,
		treeHandler,
		languageHandler,
		authMiddleware,
		cfg.Server.AllowedOrigins,
		cfg.Server.EnableHSTS,
		authLimiterMiddleware,
		apiLimiterMiddleware,
		uploadLimiterMiddleware,
	)

	engine := gin.New()
	engine.Use(gin.Recovery())
	router.Setup(engine)

	app := &App{
		cfg:          cfg,
		pool:         pool,
		engine:       engine,
		cleanupFuncs: make([]func(), 0),
	}

	app.registerCleanup(func() {
		slog.Info("Closing Redis connection")
		if err := redisClient.Close(); err != nil {
			slog.Error("Close Redis connection", "error", err)
		}
	})

	app.registerCleanup(func() {
		slog.Info("Closing database connection pool")
		pool.Close()
	})

	maintenanceWorker := worker.NewMaintenanceWorker(
		sessionRepo,
		oauthStateRepo,
		cfg.Maintenance.CleanupInterval,
	)
	maintenanceWorker.Start()
	app.registerCleanup(maintenanceWorker.Stop)

	return app, nil
}

func (a *App) registerCleanup(cleanup func()) {
	a.cleanupFuncs = append(a.cleanupFuncs, cleanup)
}

func (a *App) Close() {
	slog.Info("Application shutdown initiated", "cleanup_functions", len(a.cleanupFuncs))

	for _, fn := range slices.Backward(a.cleanupFuncs) {
		fn()
	}

	slog.Info("Application shutdown completed")
}
