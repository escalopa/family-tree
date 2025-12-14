package http

import (
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

type Router struct {
	authHandler    AuthHandler
	userHandler    UserHandler
	memberHandler  MemberHandler
	spouseHandler  SpouseHandler
	treeHandler    TreeHandler
	authMiddleware AuthMiddleware
}

func NewRouter(
	authHandler AuthHandler,
	userHandler UserHandler,
	memberHandler MemberHandler,
	spouseHandler SpouseHandler,
	treeHandler TreeHandler,
	authMiddleware AuthMiddleware,
) *Router {
	return &Router{
		authHandler:    authHandler,
		userHandler:    userHandler,
		memberHandler:  memberHandler,
		spouseHandler:  spouseHandler,
		treeHandler:    treeHandler,
		authMiddleware: authMiddleware,
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(middleware.CORS())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	auth := engine.Group("/auth")
	{
		auth.GET("/:provider", r.authHandler.GetAuthURL)
		auth.GET("/:provider/callback", r.authHandler.HandleCallback)
	}

	api := engine.Group("/api")
	api.Use(r.authMiddleware.Authenticate())
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/logout", r.authHandler.Logout)
			authGroup.POST("/logout-all", r.authHandler.LogoutAll)
		}

		userGroup := api.Group("/users")
		userGroup.Use(middleware.RequireActive())
		{
			userGroup.GET("", r.userHandler.ListUsers)
			userGroup.GET("/:user_id", r.userHandler.GetUser)
			userGroup.GET("/leaderboard", r.userHandler.GetLeaderboard)
			userGroup.GET("/score/:user_id", r.userHandler.GetScoreHistory)
			userGroup.GET("/members/:user_id", middleware.RequireRole(domain.RoleSuperAdmin), r.userHandler.GetUserChanges)

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
			memberGroup.GET("/info/:member_id", r.memberHandler.GetMember)
			memberGroup.GET("/search", r.memberHandler.SearchMembers)
			memberGroup.GET("/history", middleware.RequireRole(domain.RoleSuperAdmin), r.memberHandler.GetMemberHistory)

			memberGroup.POST("", middleware.RequireRole(domain.RoleAdmin), r.memberHandler.CreateMember)
			memberGroup.PUT("/:member_id", middleware.RequireRole(domain.RoleAdmin), r.memberHandler.UpdateMember)
			memberGroup.DELETE("/:member_id", middleware.RequireRole(domain.RoleAdmin), r.memberHandler.DeleteMember)
			memberGroup.POST("/:member_id/picture", middleware.RequireRole(domain.RoleAdmin), r.memberHandler.UploadPicture)
			memberGroup.DELETE("/:member_id/picture", middleware.RequireRole(domain.RoleAdmin), r.memberHandler.DeletePicture)
		}

		spouseGroup := api.Group("/spouses")
		spouseGroup.Use(middleware.RequireActive(), middleware.RequireRole(domain.RoleAdmin))
		{
			spouseGroup.POST("", r.spouseHandler.AddSpouse)
			spouseGroup.PUT("", r.spouseHandler.UpdateSpouse)
			spouseGroup.DELETE("", r.spouseHandler.RemoveSpouse)
		}
	}
}
