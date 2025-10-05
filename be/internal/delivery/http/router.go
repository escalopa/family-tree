package http

import (
	"github.com/escalopa/family-tree-api/internal/delivery/http/middleware"
	"github.com/gin-gonic/gin"
)

// Local interfaces for dependencies used by the router
type AuthHandler interface {
	GoogleAuth(c *gin.Context)
	GoogleCallback(c *gin.Context)
	RefreshToken(c *gin.Context)
	Logout(c *gin.Context)
	LogoutAll(c *gin.Context)
}

type UserHandler interface {
	List(c *gin.Context)
	GetByID(c *gin.Context)
	GetCurrentUser(c *gin.Context)
	GetRecentActivities(c *gin.Context)
	UpdateRole(c *gin.Context)
	UpdateActiveStatus(c *gin.Context)
	AdminLogout(c *gin.Context)
}

type MemberHandler interface {
	Search(c *gin.Context)
	Create(c *gin.Context)
	GetByID(c *gin.Context)
	Update(c *gin.Context)
	Patch(c *gin.Context)
	Delete(c *gin.Context)
	GetPicture(c *gin.Context)
	UploadPicture(c *gin.Context)
	DeletePicture(c *gin.Context)
	GetPictures(c *gin.Context)
	Rollback(c *gin.Context)
	GetHistory(c *gin.Context)
}

type SpouseHandler interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

type TreeHandler interface {
	GetTree(c *gin.Context)
	GetRelation(c *gin.Context)
}

type HistoryHandler interface {
	GetRecentActivities(c *gin.Context)
}

type ScoreHandler interface {
	GetLeaderboard(c *gin.Context)
}

type AuthMiddleware interface {
	Authenticate() gin.HandlerFunc
}

type Router struct {
	authHandler    AuthHandler
	userHandler    UserHandler
	memberHandler  MemberHandler
	spouseHandler  SpouseHandler
	treeHandler    TreeHandler
	historyHandler HistoryHandler
	scoreHandler   ScoreHandler
	authMiddleware AuthMiddleware
}

func NewRouter(
	authHandler AuthHandler,
	userHandler UserHandler,
	memberHandler MemberHandler,
	spouseHandler SpouseHandler,
	treeHandler TreeHandler,
	historyHandler HistoryHandler,
	scoreHandler ScoreHandler,
	authMiddleware AuthMiddleware,
) *Router {
	return &Router{
		authHandler:    authHandler,
		userHandler:    userHandler,
		memberHandler:  memberHandler,
		spouseHandler:  spouseHandler,
		treeHandler:    treeHandler,
		historyHandler: historyHandler,
		scoreHandler:   scoreHandler,
		authMiddleware: authMiddleware,
	}
}

func (r *Router) Setup() *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS())

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.GET("/google", r.authHandler.GoogleAuth)
			auth.GET("/google/callback", r.authHandler.GoogleCallback)
			auth.POST("/refresh", r.authHandler.RefreshToken)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(r.authMiddleware.Authenticate())
		{
			// Auth routes (authenticated)
			authGroup := protected.Group("/auth")
			{
				authGroup.POST("/logout", r.authHandler.Logout)
				authGroup.POST("/logout-all", r.authHandler.LogoutAll)
			}

			// User routes
			users := protected.Group("/users")
			{
				users.GET("", middleware.RequireViewer(), r.userHandler.List)
				users.GET("/me", r.userHandler.GetCurrentUser)
				users.GET("/me/recent-activities", r.userHandler.GetRecentActivities)
				users.GET("/:user_id", middleware.RequireViewer(), r.userHandler.GetByID)
				users.GET("/:user_id/scores", middleware.RequireViewer(), r.scoreHandler.GetLeaderboard)
				users.PUT("/:user_id/active", middleware.RequireSuperAdmin(), r.userHandler.UpdateActiveStatus)
				users.PUT("/:user_id/role", middleware.RequireSuperAdmin(), r.userHandler.UpdateRole)
				users.POST("/:user_id/logout", middleware.RequireSuperAdmin(), r.userHandler.AdminLogout)
			}

			// Member routes
			members := protected.Group("/members")
			{
				members.GET("", middleware.RequireViewer(), r.memberHandler.Search)
				members.POST("", middleware.RequireAdmin(), r.memberHandler.Create)
				members.GET("/pictures", middleware.RequireViewer(), r.memberHandler.GetPictures)
				members.GET("/:member_id", middleware.RequireViewer(), r.memberHandler.GetByID)
				members.PUT("/:member_id", middleware.RequireAdmin(), r.memberHandler.Update)
				members.PATCH("/:member_id", middleware.RequireAdmin(), r.memberHandler.Patch)
				members.DELETE("/:member_id", middleware.RequireAdmin(), r.memberHandler.Delete)
				members.GET("/:member_id/picture", middleware.RequireViewer(), r.memberHandler.GetPicture)
				members.POST("/:member_id/picture", middleware.RequireAdmin(), r.memberHandler.UploadPicture)
				members.DELETE("/:member_id/picture", middleware.RequireAdmin(), r.memberHandler.DeletePicture)
				members.POST("/:member_id/rollback", middleware.RequireAdmin(), r.memberHandler.Rollback)
				members.GET("/:member_id/history", middleware.RequireViewer(), r.memberHandler.GetHistory)
			}

			// Spouse routes
			spouses := protected.Group("/spouses")
			{
				spouses.POST("", middleware.RequireAdmin(), r.spouseHandler.Create)
				spouses.PATCH("", middleware.RequireAdmin(), r.spouseHandler.Update)
				spouses.DELETE("", middleware.RequireAdmin(), r.spouseHandler.Delete)
			}

			// Tree routes
			tree := protected.Group("/tree")
			{
				tree.GET("", middleware.RequireViewer(), r.treeHandler.GetTree)
				tree.GET("/relation", middleware.RequireViewer(), r.treeHandler.GetRelation)
			}

			// Leaderboard routes
			protected.GET("/leaderboard", middleware.RequireViewer(), r.scoreHandler.GetLeaderboard)
		}
	}

	return router
}
