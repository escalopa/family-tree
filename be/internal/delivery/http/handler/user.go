package handler

import (
	"net/http"
	"strconv"

	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/gin-gonic/gin"
)

type userHandler struct {
	userUseCase UserUseCase
}

func NewUserHandler(userUseCase UserUseCase) *userHandler {
	return &userHandler{userUseCase: userUseCase}
}

func (h *userHandler) GetUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid user_id"})
		return
	}

	user, err := h.userUseCase.GetUserWithScore(c.Request.Context(), userID)
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

func (h *userHandler) ListUsers(c *gin.Context) {
	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	if query.Limit == 0 {
		query.Limit = 20
	}

	users, nextCursor, err := h.userUseCase.ListUsers(c.Request.Context(), query.Cursor, query.Limit)
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
	if err := h.userUseCase.UpdateUserRole(c.Request.Context(), userID, req.RoleID, changedBy); err != nil {
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

	if err := h.userUseCase.UpdateUserActive(c.Request.Context(), userID, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: "active status updated"})
}

func (h *userHandler) GetLeaderboard(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	leaderboard, err := h.userUseCase.GetLeaderboard(c.Request.Context(), limit)
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

func (h *userHandler) GetScoreHistory(c *gin.Context) {
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

	if query.Limit == 0 {
		query.Limit = 20
	}

	scores, nextCursor, err := h.userUseCase.GetScoreHistory(c.Request.Context(), userID, query.Cursor, query.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	var scoresResponse []dto.ScoreHistoryResponse
	for _, s := range scores {
		scoresResponse = append(scoresResponse, dto.ScoreHistoryResponse{
			UserID:            s.UserID,
			MemberID:          s.MemberID,
			MemberArabicName:  s.MemberArabicName,
			MemberEnglishName: s.MemberEnglishName,
			FieldName:         s.FieldName,
			Points:            s.Points,
			MemberVersion:     s.MemberVersion,
			CreatedAt:         s.CreatedAt,
		})
	}

	response := dto.PaginatedScoreHistoryResponse{
		Scores:     scoresResponse,
		NextCursor: nextCursor,
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
}

func (h *userHandler) GetUserChanges(c *gin.Context) {
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

	if query.Limit == 0 {
		query.Limit = 20
	}

	changes, nextCursor, err := h.userUseCase.GetUserChanges(c.Request.Context(), userID, query.Cursor, query.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	var changesResponse []dto.HistoryResponse
	for _, h := range changes {
		changesResponse = append(changesResponse, dto.HistoryResponse{
			HistoryID:     h.HistoryID,
			MemberID:      h.MemberID,
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
