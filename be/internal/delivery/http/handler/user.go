package handler

import (
	"net/http"
	"strconv"

	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

type userHandler struct {
	userUseCase UserUseCase
}

func NewUserHandler(userUseCase UserUseCase) *userHandler {
	return &userHandler{userUseCase: userUseCase}
}

// GetUser godoc
// @Summary Get user by ID
// @Description Returns a user with their total score
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "User ID"
// @Success 200 {object} dto.Response{data=dto.UserResponse}
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /api/users/{user_id} [get]
func (h *userHandler) Get(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid user_id"})
		return
	}

	user, err := h.userUseCase.GetWithScore(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Response{Success: false, Error: err.Error()})
		return
	}

	response := dto.UserResponse{
		UserID:     user.UserID,
		FullName:   user.FullName,
		Email:      user.Email,
		Avatar:     user.Avatar,
		RoleID:     user.RoleID,
		IsActive:   user.IsActive,
		TotalScore: &user.TotalScore,
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
}

// ListUsers godoc
// @Summary List all users
// @Description Returns a paginated list of users with optional filters
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param search query string false "Search by name or email"
// @Param role_id query int false "Filter by role ID"
// @Param is_active query bool false "Filter by active status"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Number of items to return (1-100)" default(20)
// @Success 200 {object} dto.Response{data=dto.PaginatedUsersResponse}
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /api/users [get]
func (h *userHandler) List(c *gin.Context) {
	var query dto.UserFilterQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	filter := domain.UserFilter{
		Search:   query.Search,
		RoleID:   query.RoleID,
		IsActive: query.IsActive,
	}

	users, nextCursor, err := h.userUseCase.List(c.Request.Context(), filter, query.Cursor, query.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	var usersResponse []dto.UserResponse
	for _, u := range users {
		usersResponse = append(usersResponse, dto.UserResponse{
			UserID:   u.UserID,
			FullName: u.FullName,
			Email:    u.Email,
			Avatar:   u.Avatar,
			RoleID:   u.RoleID,
			IsActive: u.IsActive,
		})
	}

	response := dto.PaginatedUsersResponse{
		Users:      usersResponse,
		NextCursor: nextCursor,
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
}

func (h *userHandler) UpdateRole(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid user_id"})
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	changedBy := middleware.GetUserID(c)
	if err := h.userUseCase.UpdateRole(c.Request.Context(), userID, req.RoleID, changedBy); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: "role updated"})
}

func (h *userHandler) UpdateActive(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid user_id"})
		return
	}

	var req dto.UpdateActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	if err := h.userUseCase.UpdateActive(c.Request.Context(), userID, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: "active status updated"})
}

// GetLeaderboard godoc
// @Summary Get user leaderboard
// @Description Returns the top users ranked by total score
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of top users to return" default(10)
// @Success 200 {object} dto.Response{data=dto.LeaderboardResponse}
// @Failure 401 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /api/users/leaderboard [get]
func (h *userHandler) ListLeaderboard(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	leaderboard, err := h.userUseCase.ListLeaderboard(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	var response dto.LeaderboardResponse
	for _, u := range leaderboard {
		response.Users = append(response.Users, dto.UserScore{
			UserID:     u.UserID,
			FullName:   u.FullName,
			Avatar:     u.Avatar,
			TotalScore: u.TotalScore,
			Rank:       u.Rank,
		})
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
}

func (h *userHandler) ListScoreHistory(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid user_id"})
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	scores, nextCursor, err := h.userUseCase.ListScoreHistory(c.Request.Context(), userID, query.Cursor, query.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	preferredLang := middleware.GetPreferredLanguage(c)

	var scoresResponse []dto.ScoreHistoryResponse
	for _, s := range scores {
		memberName := extractName(s.MemberNames, preferredLang)
		scoresResponse = append(scoresResponse, dto.ScoreHistoryResponse{
			UserID:        s.UserID,
			MemberID:      s.MemberID,
			MemberName:    memberName,
			FieldName:     s.FieldName,
			Points:        s.Points,
			MemberVersion: s.MemberVersion,
			CreatedAt:     s.CreatedAt,
		})
	}

	response := dto.PaginatedScoreHistoryResponse{
		Scores:     scoresResponse,
		NextCursor: nextCursor,
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
}

func (h *userHandler) ListChanges(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid user_id"})
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	changes, nextCursor, err := h.userUseCase.ListChanges(c.Request.Context(), userID, query.Cursor, query.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	preferredLang := middleware.GetPreferredLanguage(c)

	var changesResponse []dto.HistoryResponse
	for _, h := range changes {
		memberName := extractName(h.MemberNames, preferredLang)
		changesResponse = append(changesResponse, dto.HistoryResponse{
			HistoryID:     h.HistoryID,
			MemberID:      h.MemberID,
			MemberName:    memberName,
			UserID:        h.UserID,
			UserFullName:  h.UserFullName,
			UserEmail:     h.UserEmail,
			ChangedAt:     h.ChangedAt,
			ChangeType:    h.ChangeType,
			OldValues:     h.OldValues,
			NewValues:     h.NewValues,
			MemberVersion: h.MemberVersion,
		})
	}

	response := dto.PaginatedHistoryResponse{
		History:    changesResponse,
		NextCursor: nextCursor,
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
}
