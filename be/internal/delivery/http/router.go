package http

import (
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	authHandler               AuthHandler
	userHandler               UserHandler
	memberHandler             MemberHandler
	spouseHandler             SpouseHandler
	treeHandler               TreeHandler
	familyTreeHandler         FamilyTreeHandler
	languageHandler           LanguageHandler
	authMiddleware            AuthMiddleware
	allowedOrigins            []string
	enableHSTS                bool
	authRateLimitMiddleware   RateLimitMiddleware
	apiRateLimitMiddleware    RateLimitMiddleware
	uploadRateLimitMiddleware RateLimitMiddleware
}

func NewRouter(
	authHandler AuthHandler,
	userHandler UserHandler,
	memberHandler MemberHandler,
	spouseHandler SpouseHandler,
	treeHandler TreeHandler,
	familyTreeHandler FamilyTreeHandler,
	languageHandler LanguageHandler,
	authMiddleware AuthMiddleware,
	allowedOrigins []string,
	enableHSTS bool,
	authRateLimitMiddleware RateLimitMiddleware,
	apiRateLimitMiddleware RateLimitMiddleware,
	uploadRateLimitMiddleware RateLimitMiddleware,
) *Router {
	return &Router{
		authHandler:               authHandler,
		userHandler:               userHandler,
		memberHandler:             memberHandler,
		spouseHandler:             spouseHandler,
		treeHandler:               treeHandler,
		familyTreeHandler:         familyTreeHandler,
		languageHandler:           languageHandler,
		authMiddleware:            authMiddleware,
		allowedOrigins:            allowedOrigins,
		enableHSTS:                enableHSTS,
		authRateLimitMiddleware:   authRateLimitMiddleware,
		apiRateLimitMiddleware:    apiRateLimitMiddleware,
		uploadRateLimitMiddleware: uploadRateLimitMiddleware,
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(middleware.SecurityHeaders(r.enableHSTS))
	engine.Use(middleware.CORS(r.allowedOrigins))
	engine.Use(middleware.LanguageMiddleware())

	engine.GET("/swagger/", ginSwagger.WrapHandler(swaggerFiles.Handler))

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	engine.GET("/public/trees/:token", r.familyTreeHandler.GetPublicTree)

	auth := engine.Group("/auth")
	auth.Use(r.authRateLimitMiddleware.RateLimit())
	{
		auth.GET("/providers", r.authHandler.GetProviders)
		auth.GET("/:provider", r.authHandler.GetAuthURL)
		auth.GET("/:provider/callback", r.authHandler.HandleCallback)
	}

	api := engine.Group("/api")
	api.Use(
		r.authMiddleware.Authenticate(),
		r.apiRateLimitMiddleware.RateLimit(),
	)
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

			userGroup.PUT("/:user_id", middleware.RequireRole(domain.RoleSuperAdmin), r.userHandler.Update)
		}

		familyTreeGroup := api.Group("/family-trees")
		familyTreeGroup.Use(middleware.RequireActive())
		{
			familyTreeGroup.GET("", r.familyTreeHandler.List)
			familyTreeGroup.POST("", r.familyTreeHandler.Create)
			familyTreeGroup.GET("/invitations", r.familyTreeHandler.ListMyInvitations)
			familyTreeGroup.POST("/invitations/:invitation_id/accept", r.familyTreeHandler.AcceptInvitation)
			familyTreeGroup.POST("/invitations/:invitation_id/decline", r.familyTreeHandler.DeclineInvitation)
			familyTreeGroup.GET("/:tree_id", r.familyTreeHandler.Get)
			familyTreeGroup.GET("/:tree_id/tree", r.treeHandler.GetTree)
			familyTreeGroup.GET("/:tree_id/tree/graph", r.treeHandler.GetGraph)
			familyTreeGroup.GET("/:tree_id/tree/relation", r.treeHandler.GetRelation)
			familyTreeGroup.GET("/:tree_id/tree/graph/relation", r.treeHandler.GetRelationGraph)
			familyTreeGroup.GET("/:tree_id/members", r.memberHandler.List)
			familyTreeGroup.GET("/:tree_id/members/search", r.memberHandler.List)
			familyTreeGroup.GET("/:tree_id/members/history", middleware.RequireRole(domain.RoleSuperAdmin), r.memberHandler.ListHistory)
			familyTreeGroup.GET("/:tree_id/members/:member_id", r.memberHandler.Get)
			familyTreeGroup.GET("/:tree_id/members/:member_id/picture", middleware.RequireRole(domain.RoleAdmin), r.memberHandler.GetPicture)
			familyTreeGroup.POST("/:tree_id/members", middleware.RequireRole(domain.RoleAdmin), r.memberHandler.Create)
			familyTreeGroup.POST("/:tree_id/members/:member_id/rollback", middleware.RequireRole(domain.RoleSuperAdmin), r.memberHandler.Rollback)
			familyTreeGroup.PUT("/:tree_id/members/:member_id", middleware.RequireRole(domain.RoleAdmin), r.memberHandler.Update)
			familyTreeGroup.DELETE("/:tree_id/members/:member_id", middleware.RequireRole(domain.RoleSuperAdmin), r.memberHandler.Delete)
			familyTreeGroup.POST("/:tree_id/members/:member_id/picture", r.uploadRateLimitMiddleware.RateLimit(), middleware.RequireRole(domain.RoleAdmin), r.memberHandler.UploadPicture)
			familyTreeGroup.DELETE("/:tree_id/members/:member_id/picture", middleware.RequireRole(domain.RoleAdmin), r.memberHandler.DeletePicture)
			familyTreeGroup.POST("/:tree_id/spouses", middleware.RequireRole(domain.RoleAdmin), r.spouseHandler.Create)
			familyTreeGroup.PUT("/:tree_id/spouses/:spouse_id", middleware.RequireRole(domain.RoleAdmin), r.spouseHandler.Update)
			familyTreeGroup.DELETE("/:tree_id/spouses/:spouse_id", middleware.RequireRole(domain.RoleAdmin), r.spouseHandler.Delete)
			familyTreeGroup.POST("/:tree_id/invitations", r.familyTreeHandler.Invite)
			familyTreeGroup.GET("/:tree_id/invitations", r.familyTreeHandler.ListInvitations)
			familyTreeGroup.POST("/:tree_id/share-links", r.familyTreeHandler.CreateShareLink)
			familyTreeGroup.GET("/:tree_id/share-links", r.familyTreeHandler.ListShareLinks)
			familyTreeGroup.DELETE("/:tree_id/share-links/:share_id", r.familyTreeHandler.RevokeShareLink)
		}

		languageGroup := api.Group("/languages")
		{
			languageGroup.GET("", r.languageHandler.List)
			languageGroup.GET("/:code", r.languageHandler.Get)
			languageGroup.PATCH("/:code/toggle", middleware.RequireActive(), middleware.RequireRole(domain.RoleSuperAdmin), r.languageHandler.ToggleActive)
			languageGroup.PUT("/order", middleware.RequireActive(), middleware.RequireRole(domain.RoleSuperAdmin), r.languageHandler.UpdateOrder)
		}

		userPrefsGroup := api.Group("/users/me/preferences")
		userPrefsGroup.Use(middleware.RequireActive())
		{
			userPrefsGroup.GET("/languages", r.languageHandler.GetPreference)
			userPrefsGroup.PUT("/languages", r.languageHandler.UpdatePreference)
		}
	}
}
