package http

import "github.com/gin-gonic/gin"

// Handler interfaces used by the HTTP router

type AuthHandler interface {
	GetAuthURL(c *gin.Context)
	HandleCallback(c *gin.Context)
	RefreshToken(c *gin.Context)
	Logout(c *gin.Context)
	LogoutAll(c *gin.Context)
}

type UserHandler interface {
	GetUser(c *gin.Context)
	ListUsers(c *gin.Context)
	UpdateRole(c *gin.Context)
	UpdateActive(c *gin.Context)
	GetLeaderboard(c *gin.Context)
	GetScoreHistory(c *gin.Context)
	GetUserChanges(c *gin.Context)
}

type MemberHandler interface {
	CreateMember(c *gin.Context)
	UpdateMember(c *gin.Context)
	DeleteMember(c *gin.Context)
	GetMember(c *gin.Context)
	SearchMembers(c *gin.Context)
	GetMemberHistory(c *gin.Context)
	UploadPicture(c *gin.Context)
	DeletePicture(c *gin.Context)
}

type SpouseHandler interface {
	AddSpouse(c *gin.Context)
	UpdateSpouse(c *gin.Context)
	RemoveSpouse(c *gin.Context)
}

type TreeHandler interface {
	GetTree(c *gin.Context)
	GetRelation(c *gin.Context)
}

// AuthMiddleware interface used by the HTTP router

type AuthMiddleware interface {
	Authenticate() gin.HandlerFunc
}
