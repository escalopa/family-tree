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
	SearchMemberInfo(c *gin.Context)
	GetMemberHistory(c *gin.Context)
	UploadPicture(c *gin.Context)
	DeletePicture(c *gin.Context)
	GetPicture(c *gin.Context)
}

type SpouseHandler interface {
	AddSpouse(c *gin.Context)
	UpdateSpouse(c *gin.Context)
	UpdateSpouseByID(c *gin.Context)
	RemoveSpouse(c *gin.Context)
}

type TreeHandler interface {
	GetTree(c *gin.Context)
	GetRelation(c *gin.Context)
}

type LanguageHandler interface {
	GetLanguages(c *gin.Context)
	GetLanguage(c *gin.Context)
	CreateLanguage(c *gin.Context)
	UpdateLanguage(c *gin.Context)
	GetUserLanguagePreference(c *gin.Context)
	UpdateUserLanguagePreference(c *gin.Context)
}

type AuthMiddleware interface {
	Authenticate() gin.HandlerFunc
}
