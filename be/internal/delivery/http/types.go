package http

import "github.com/gin-gonic/gin"

type AuthHandler interface {
	GetProviders(c *gin.Context)
	GetAuthURL(c *gin.Context)
	HandleCallback(c *gin.Context)
	GetCurrentUser(c *gin.Context)
	Logout(c *gin.Context)
	LogoutAll(c *gin.Context)
}

type UserHandler interface {
	Get(c *gin.Context)
	List(c *gin.Context)
	UpdateRole(c *gin.Context)
	UpdateActive(c *gin.Context)
	ListLeaderboard(c *gin.Context)
	ListScoreHistory(c *gin.Context)
	ListChanges(c *gin.Context)
}

type MemberHandler interface {
	Get(c *gin.Context)
	List(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	ListHistory(c *gin.Context)
	UploadPicture(c *gin.Context)
	DeletePicture(c *gin.Context)
	GetPicture(c *gin.Context)
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

type LanguageHandler interface {
	Get(c *gin.Context)
	List(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	GetPreference(c *gin.Context)
	UpdatePreference(c *gin.Context)
}

type AuthMiddleware interface {
	Authenticate() gin.HandlerFunc
}
