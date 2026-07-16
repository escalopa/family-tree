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
	Update(c *gin.Context)
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
	Rollback(c *gin.Context)
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
	GetGraph(c *gin.Context)
	GetRelationGraph(c *gin.Context)
}

type FamilyTreeHandler interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	Get(c *gin.Context)
	ListMyInvitations(c *gin.Context)
	Invite(c *gin.Context)
	ListInvitations(c *gin.Context)
	AcceptInvitation(c *gin.Context)
	DeclineInvitation(c *gin.Context)
	CreateShareLink(c *gin.Context)
	ListShareLinks(c *gin.Context)
	RevokeShareLink(c *gin.Context)
	GetPublicTree(c *gin.Context)
}

type LanguageHandler interface {
	Get(c *gin.Context)
	List(c *gin.Context)
	ToggleActive(c *gin.Context)
	GetPreference(c *gin.Context)
	UpdatePreference(c *gin.Context)
	UpdateOrder(c *gin.Context)
}

type AuthMiddleware interface {
	Authenticate() gin.HandlerFunc
}

type RateLimitMiddleware interface {
	RateLimit() gin.HandlerFunc
}
