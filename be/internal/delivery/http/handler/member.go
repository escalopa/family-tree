package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

type memberHandler struct {
	memberUseCase MemberUseCase
}

func NewMemberHandler(memberUseCase MemberUseCase) *memberHandler {
	return &memberHandler{memberUseCase: memberUseCase}
}

func (h *memberHandler) CreateMember(c *gin.Context) {
	var req dto.CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	member := &domain.Member{
		ArabicName:  req.ArabicName,
		EnglishName: req.EnglishName,
		Gender:      req.Gender,
		DateOfBirth: req.DateOfBirth,
		DateOfDeath: req.DateOfDeath,
		FatherID:    req.FatherID,
		MotherID:    req.MotherID,
		Nicknames:   req.Nicknames,
		Profession:  req.Profession,
	}

	userID := middleware.GetUserID(c)
	if err := h.memberUseCase.CreateMember(c.Request.Context(), member, userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.Response{Success: true, Data: gin.H{"member_id": member.MemberID, "version": member.Version}})
}

func (h *memberHandler) UpdateMember(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("member_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid member_id"})
		return
	}

	var req dto.UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	member := &domain.Member{
		MemberID:    memberID,
		ArabicName:  req.ArabicName,
		EnglishName: req.EnglishName,
		Gender:      req.Gender,
		DateOfBirth: req.DateOfBirth,
		DateOfDeath: req.DateOfDeath,
		FatherID:    req.FatherID,
		MotherID:    req.MotherID,
		Nicknames:   req.Nicknames,
		Profession:  req.Profession,
	}

	userID := middleware.GetUserID(c)
	if err := h.memberUseCase.UpdateMember(c.Request.Context(), member, req.Version, userID); err != nil {
		c.JSON(http.StatusConflict, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: gin.H{"version": member.Version}})
}

func (h *memberHandler) DeleteMember(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("member_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid member_id"})
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.memberUseCase.DeleteMember(c.Request.Context(), memberID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: "member deleted"})
}

func (h *memberHandler) GetMember(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("member_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid member_id"})
		return
	}

	member, err := h.memberUseCase.GetMemberByID(c.Request.Context(), memberID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Response{Success: false, Error: err.Error()})
		return
	}

	userRole := middleware.GetUserRole(c)
	computed := h.memberUseCase.ComputeMemberWithExtras(c.Request.Context(), member, userRole)

	response := dto.MemberResponse{
		MemberID:        computed.MemberID,
		ArabicName:      computed.ArabicName,
		EnglishName:     computed.EnglishName,
		Gender:          computed.Gender,
		Picture:         computed.Picture,
		DateOfBirth:     computed.DateOfBirth,
		DateOfDeath:     computed.DateOfDeath,
		FatherID:        computed.FatherID,
		MotherID:        computed.MotherID,
		Nicknames:       computed.Nicknames,
		Profession:      computed.Profession,
		Version:         computed.Version,
		ArabicFullName:  computed.ArabicFullName,
		EnglishFullName: computed.EnglishFullName,
		Age:             computed.Age,
		GenerationLevel: computed.GenerationLevel,
		IsMarried:       computed.IsMarried,
		Spouses:         computed.Spouses,
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
}

func (h *memberHandler) SearchMembers(c *gin.Context) {
	var query dto.MemberSearchQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	// At least one filter must be provided
	if query.ArabicName == nil && query.EnglishName == nil && query.Gender == nil && query.Married == nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "at least one filter required"})
		return
	}

	filter := domain.MemberFilter{
		ArabicName:  query.ArabicName,
		EnglishName: query.EnglishName,
		Gender:      query.Gender,
	}
	if query.Married != nil {
		married := *query.Married == 1
		filter.IsMarried = &married
	}

	// Set default limit if not provided
	limit := 20
	if query.Limit > 0 {
		limit = query.Limit
	}

	members, nextCursor, err := h.memberUseCase.SearchMembers(c.Request.Context(), filter, query.Cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	userRole := middleware.GetUserRole(c)
	var membersResponse []dto.MemberResponse
	for _, m := range members {
		computed := h.memberUseCase.ComputeMemberWithExtras(c.Request.Context(), m, userRole)
		membersResponse = append(membersResponse, dto.MemberResponse{
			MemberID:    computed.MemberID,
			ArabicName:  computed.ArabicName,
			EnglishName: computed.EnglishName,
			Gender:      computed.Gender,
			Picture:     computed.Picture,
			DateOfBirth: computed.DateOfBirth,
			DateOfDeath: computed.DateOfDeath,
			FatherID:    computed.FatherID,
			MotherID:    computed.MotherID,
			Nicknames:   computed.Nicknames,
			Profession:  computed.Profession,
			Version:     computed.Version,
			IsMarried:   computed.IsMarried,
			Spouses:     computed.Spouses,
		})
	}

	response := dto.PaginatedMembersResponse{
		Members:    membersResponse,
		NextCursor: nextCursor,
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
}

func (h *memberHandler) GetMemberHistory(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Query("member_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid member_id"})
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

	history, nextCursor, err := h.memberUseCase.GetMemberHistory(c.Request.Context(), memberID, query.Cursor, query.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	var historyResponse []dto.HistoryResponse
	for _, h := range history {
		historyResponse = append(historyResponse, dto.HistoryResponse{
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
		History:    historyResponse,
		NextCursor: nextCursor,
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
}

func (h *memberHandler) UploadPicture(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("member_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid member_id"})
		return
	}

	file, header, err := c.Request.FormFile("picture")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "missing picture file"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: "failed to read file"})
		return
	}

	userID := middleware.GetUserID(c)
	pictureURL, err := h.memberUseCase.UploadPicture(c.Request.Context(), memberID, data, header.Filename, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: gin.H{"picture_url": pictureURL}})
}

func (h *memberHandler) DeletePicture(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("member_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid member_id"})
		return
	}

	if err := h.memberUseCase.DeletePicture(c.Request.Context(), memberID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: "picture deleted"})
}
