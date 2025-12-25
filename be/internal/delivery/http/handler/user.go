package handler

import (
	"strconv"

	"github.com/escalopa/family-tree/internal/delivery"
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
	var uri dto.UserIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	user, err := h.userUseCase.GetWithScore(c.Request.Context(), uri.UserID)
	if err != nil {
		delivery.Error(c, err)
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
	delivery.SuccessWithData(c, response)
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
		delivery.Error(c, err)
		return
	}

	filter := domain.UserFilter{
		Search:   query.Search,
		RoleID:   query.RoleID,
		IsActive: query.IsActive,
	}

	users, nextCursor, err := h.userUseCase.List(c.Request.Context(), filter, query.Cursor, query.Limit)
	if err != nil {
		delivery.Error(c, err)
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

	delivery.SuccessWithData(c, response)
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user role and/or active status
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "User ID"
// @Param request body dto.UpdateUserRequest true "Update user request"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /api/users/{user_id} [put]
func (h *userHandler) Update(c *gin.Context) {
	var uri dto.UserIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		delivery.Error(c, err)
		return
	}

	if req.RoleID == nil && req.IsActive == nil {
		delivery.Error(c, domain.NewValidationError("error.validation.at_least_one_field_required"))
		return
	}

	changedBy := middleware.GetUserID(c)
	if err := h.userUseCase.Update(c.Request.Context(), uri.UserID, req.RoleID, req.IsActive, changedBy); err != nil {
		delivery.Error(c, err)
		return
	}

	delivery.Success(c, "success.user.updated", nil)
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
		delivery.Error(c, err)
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

	delivery.SuccessWithData(c, response)
}

func (h *userHandler) ListScoreHistory(c *gin.Context) {
	var uri dto.UserIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		delivery.Error(c, err)
		return
	}

	scores, nextCursor, err := h.userUseCase.ListScoreHistory(c.Request.Context(), uri.UserID, query.Cursor, query.Limit)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	preferredLang := middleware.GetPreferredLanguage(c)

	var scoresResponse []dto.ScoreHistoryResponse
	for _, s := range scores {
		scoresResponse = append(scoresResponse, dto.ScoreHistoryResponse{
			UserID:        s.UserID,
			MemberID:      s.MemberID,
			MemberName:    extractName(s.MemberNames, preferredLang),
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

	delivery.SuccessWithData(c, response)
}

func (h *userHandler) ListChanges(c *gin.Context) {
	var uri dto.UserIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		delivery.Error(c, err)
		return
	}

	changes, nextCursor, err := h.userUseCase.ListChanges(c.Request.Context(), uri.UserID, query.Cursor, query.Limit)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	preferredLang := middleware.GetPreferredLanguage(c)

	var changesResponse []dto.HistoryResponse
	for _, h := range changes {
		changesResponse = append(changesResponse, dto.HistoryResponse{
			HistoryID:     h.HistoryID,
			MemberID:      h.MemberID,
			MemberName:    extractName(h.MemberNames, preferredLang),
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

	delivery.SuccessWithData(c, response)
}
