package http

import (
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	authHandler     AuthHandler
	userHandler     UserHandler
	memberHandler   MemberHandler
	spouseHandler   SpouseHandler
	treeHandler     TreeHandler
	languageHandler LanguageHandler
	authMiddleware  AuthMiddleware
	allowedOrigins  []string
}

func NewRouter(
	authHandler AuthHandler,
	userHandler UserHandler,
	memberHandler MemberHandler,
	spouseHandler SpouseHandler,
	treeHandler TreeHandler,
	languageHandler LanguageHandler,
	authMiddleware AuthMiddleware,
	allowedOrigins []string,
) *Router {
	return &Router{
		authHandler:     authHandler,
		userHandler:     userHandler,
		memberHandler:   memberHandler,
		spouseHandler:   spouseHandler,
		treeHandler:     treeHandler,
		languageHandler: languageHandler,
		authMiddleware:  authMiddleware,
		allowedOrigins:  allowedOrigins,
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(middleware.CORS(r.allowedOrigins))

	engine.GET("/swagger/", ginSwagger.WrapHandler(swaggerFiles.Handler))

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	auth := engine.Group("/auth")
	{
		auth.GET("/providers", r.authHandler.GetProviders)
		auth.GET("/:provider", r.authHandler.GetAuthURL)
		auth.GET("/:provider/callback", r.authHandler.HandleCallback)
	}

	api := engine.Group("/api")
	api.Use(r.authMiddleware.Authenticate())
	{
		authGroup := api.Group("/auth")
		{
			authGroup.GET("/me", r.authHandler.GetCurrentUser)
			authGroup.POST("/logout", r.authHandler.Logout)
			authGroup.POST("/logout-all", r.authHandler.LogoutAll)
		}

		userGroup := api.Group("/users")
		userGroup.Use(middleware.RequireActive())
		{
			userGroup.GET("", r.userHandler.List)
			userGroup.GET("/leaderboard", r.userHandler.ListLeaderboard)
			userGroup.GET("/score/:user_id", r.userHandler.ListScoreHistory)
			userGroup.GET("/members/:user_id", middleware.RequireRole(domain.RoleAdmin), r.userHandler.ListChanges)
			userGroup.GET("/:user_id", r.userHandler.Get)

			userGroup.PUT("/:user_id/role", middleware.RequireRole(domain.RoleSuperAdmin), r.userHandler.UpdateRole)
			userGroup.PUT("/:user_id/active", middleware.RequireRole(domain.RoleSuperAdmin), r.userHandler.UpdateActive)
		}

		treeGroup := api.Group("/tree")
		treeGroup.Use(middleware.RequireActive())
		{
			treeGroup.GET("", r.treeHandler.GetTree)
			treeGroup.GET("/relation", r.treeHandler.GetRelation)
		}

		memberGroup := api.Group("/members")
		memberGroup.Use(middleware.RequireActive())
		{
			memberGroup.GET("", r.memberHandler.List)
			memberGroup.GET("/history", middleware.RequireRole(domain.RoleSuperAdmin), r.memberHandler.ListHistory)
			memberGroup.GET("/:member_id", r.memberHandler.Get)
			memberGroup.GET("/:member_id/picture", middleware.RequireRole(domain.RoleAdmin), r.memberHandler.GetPicture)

			memberGroup.POST("", middleware.RequireRole(domain.RoleAdmin), r.memberHandler.Create)
			memberGroup.PUT("/:member_id", middleware.RequireRole(domain.RoleAdmin), r.memberHandler.Update)
			memberGroup.DELETE("/:member_id", middleware.RequireRole(domain.RoleSuperAdmin), r.memberHandler.Delete)
			memberGroup.POST("/:member_id/picture", middleware.RequireRole(domain.RoleAdmin), r.memberHandler.UploadPicture)
			memberGroup.DELETE("/:member_id/picture", middleware.RequireRole(domain.RoleAdmin), r.memberHandler.DeletePicture)
		}

		spouseGroup := api.Group("/spouses")
		spouseGroup.Use(middleware.RequireActive(), middleware.RequireRole(domain.RoleAdmin))
		{
			spouseGroup.POST("", r.spouseHandler.Create)
			spouseGroup.PUT("", r.spouseHandler.Update)
			spouseGroup.DELETE("", r.spouseHandler.Delete)
		}

		languageGroup := api.Group("/languages")
		{
			languageGroup.GET("", r.languageHandler.List)
			languageGroup.GET("/:code", r.languageHandler.Get)

			languageGroup.POST("", middleware.RequireActive(), middleware.RequireRole(domain.RoleSuperAdmin), r.languageHandler.Create)
			languageGroup.PUT("/:code", middleware.RequireActive(), middleware.RequireRole(domain.RoleSuperAdmin), r.languageHandler.Update)
		}

		userPrefsGroup := api.Group("/users/me/preferences")
		userPrefsGroup.Use(middleware.RequireActive())
		{
			userPrefsGroup.GET("/languages", r.languageHandler.GetPreference)
			userPrefsGroup.PUT("/languages", r.languageHandler.UpdatePreference)
		}
	}
}
