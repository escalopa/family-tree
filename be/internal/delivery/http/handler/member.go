package handler

import (
	"io"
	"net/http"
	"strconv"
	"time"

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

// CreateMember godoc
// @Summary Create a new family member
// @Description Creates a new member in the family tree (requires Admin role)
// @Tags members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param member body dto.CreateMemberRequest true "Member details"
// @Success 201 {object} dto.Response{data=object{member_id=int,version=int}}
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /api/members [post]
func (h *memberHandler) CreateMember(c *gin.Context) {
	var req dto.CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	var dateOfBirth, dateOfDeath *time.Time
	if req.DateOfBirth != nil {
		dateOfBirth = req.DateOfBirth.ToTimePtr()
	}
	if req.DateOfDeath != nil {
		dateOfDeath = req.DateOfDeath.ToTimePtr()
	}

	// Ensure nicknames is never nil, use empty array instead
	nicknames := req.Nicknames
	if nicknames == nil {
		nicknames = []string{}
	}

	member := &domain.Member{
		ArabicName:  req.ArabicName,
		EnglishName: req.EnglishName,
		Gender:      req.Gender,
		DateOfBirth: dateOfBirth,
		DateOfDeath: dateOfDeath,
		FatherID:    req.FatherID,
		MotherID:    req.MotherID,
		Nicknames:   nicknames,
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

	var dateOfBirth, dateOfDeath *time.Time
	if req.DateOfBirth != nil {
		dateOfBirth = req.DateOfBirth.ToTimePtr()
	}
	if req.DateOfDeath != nil {
		dateOfDeath = req.DateOfDeath.ToTimePtr()
	}

	// Ensure nicknames is never nil, use empty array instead
	nicknames := req.Nicknames
	if nicknames == nil {
		nicknames = []string{}
	}

	member := &domain.Member{
		MemberID:    memberID,
		ArabicName:  req.ArabicName,
		EnglishName: req.EnglishName,
		Gender:      req.Gender,
		DateOfBirth: dateOfBirth,
		DateOfDeath: dateOfDeath,
		FatherID:    req.FatherID,
		MotherID:    req.MotherID,
		Nicknames:   nicknames,
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

// GetMember godoc
// @Summary Get member by ID
// @Description Returns detailed information about a family member including computed fields
// @Tags members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param member_id path int true "Member ID"
// @Success 200 {object} dto.Response{data=dto.MemberResponse}
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /api/members/info/{member_id} [get]
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

	spousesDTO := make([]dto.SpouseInfo, len(computed.Spouses))
	for i, spouse := range computed.Spouses {
		spousesDTO[i] = dto.SpouseInfo{
			SpouseID:     spouse.SpouseID,
			MemberID:     spouse.MemberID,
			ArabicName:   spouse.ArabicName,
			EnglishName:  spouse.EnglishName,
			Gender:       spouse.Gender,
			Picture:      spouse.Picture,
			MarriageDate: dto.FromTimePtr(spouse.MarriageDate),
			DivorceDate:  dto.FromTimePtr(spouse.DivorceDate),
			MarriedYears: dto.CalculateMarriedYears(spouse.MarriageDate, spouse.DivorceDate),
		}
	}

	var fatherInfo, motherInfo *dto.MemberInfo
	if computed.FatherID != nil {
		father, err := h.memberUseCase.GetMemberByID(c.Request.Context(), *computed.FatherID)
		if err == nil {
			fatherInfo = &dto.MemberInfo{
				MemberID:    father.MemberID,
				ArabicName:  father.ArabicName,
				EnglishName: father.EnglishName,
				Picture:     father.Picture,
			}
		}
	}
	if computed.MotherID != nil {
		mother, err := h.memberUseCase.GetMemberByID(c.Request.Context(), *computed.MotherID)
		if err == nil {
			motherInfo = &dto.MemberInfo{
				MemberID:    mother.MemberID,
				ArabicName:  mother.ArabicName,
				EnglishName: mother.EnglishName,
				Picture:     mother.Picture,
			}
		}
	}

	var childrenInfo []dto.MemberInfo
	children, err := h.memberUseCase.GetChildrenByParentID(c.Request.Context(), memberID)
	if err == nil {
		for _, child := range children {
			childrenInfo = append(childrenInfo, dto.MemberInfo{
				MemberID:    child.MemberID,
				ArabicName:  child.ArabicName,
				EnglishName: child.EnglishName,
				Picture:     child.Picture,
			})
		}
	}

	var siblingsInfo []dto.MemberInfo
	siblings, err := h.memberUseCase.GetSiblingsByMemberID(c.Request.Context(), memberID)
	if err == nil {
		for _, sibling := range siblings {
			siblingsInfo = append(siblingsInfo, dto.MemberInfo{
				MemberID:    sibling.MemberID,
				ArabicName:  sibling.ArabicName,
				EnglishName: sibling.EnglishName,
				Picture:     sibling.Picture,
			})
		}
	}

	response := dto.MemberResponse{
		MemberID:        computed.MemberID,
		ArabicName:      computed.ArabicName,
		EnglishName:     computed.EnglishName,
		Gender:          computed.Gender,
		Picture:         computed.Picture,
		DateOfBirth:     dto.FromTimePtr(computed.DateOfBirth),
		DateOfDeath:     dto.FromTimePtr(computed.DateOfDeath),
		FatherID:        computed.FatherID,
		MotherID:        computed.MotherID,
		Father:          fatherInfo,
		Mother:          motherInfo,
		Nicknames:       computed.Nicknames,
		Profession:      computed.Profession,
		Version:         computed.Version,
		ArabicFullName:  computed.ArabicFullName,
		EnglishFullName: computed.EnglishFullName,
		Age:             computed.Age,
		GenerationLevel: computed.GenerationLevel,
		IsMarried:       computed.IsMarried,
		Spouses:         spousesDTO,
		Children:        childrenInfo,
		Siblings:        siblingsInfo,
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
}

// SearchMembers godoc
// @Summary Search family members
// @Description Search for members by name, gender, or marital status (at least one filter required)
// @Tags members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param arabic_name query string false "Arabic name to search for"
// @Param english_name query string false "English name to search for"
// @Param gender query string false "Gender filter (male/female)"
// @Param married query int false "Marital status (0=unmarried, 1=married)"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Number of items to return (1-100)" default(20)
// @Success 200 {object} dto.Response{data=dto.PaginatedMembersResponse}
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /api/members/search [get]
func (h *memberHandler) SearchMembers(c *gin.Context) {
	var query dto.MemberSearchQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	filter := domain.MemberFilter{
		Name:   query.Name,
		Gender: query.Gender,
	}
	if query.Married != nil {
		married := *query.Married == 1
		filter.IsMarried = &married
	}

	members, nextCursor, err := h.memberUseCase.SearchMembers(c.Request.Context(), filter, query.Cursor, query.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	var membersResponse []dto.MemberListItem
	for _, m := range members {
		membersResponse = append(membersResponse, dto.MemberListItem{
			MemberID:    m.MemberID,
			ArabicName:  m.ArabicName,
			EnglishName: m.EnglishName,
			Gender:      m.Gender,
			Picture:     m.Picture,
			DateOfBirth: dto.FromTimePtr(m.DateOfBirth),
			DateOfDeath: dto.FromTimePtr(m.DateOfDeath),
			IsMarried:   m.IsMarried, // Now comes from the query result
		})
	}

	response := dto.PaginatedMembersResponse{
		Members:    membersResponse,
		NextCursor: nextCursor,
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: response})
}

// SearchMemberInfo godoc
// @Summary Search for member info
// @Description Search for members by name filtered by gender (for parent/spouse selection)
// @Tags members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query (name in Arabic or English)"
// @Param gender query string true "Required gender filter (M or F)"
// @Param limit query int false "Number of results (max 20)" default(10)
// @Success 200 {object} dto.Response{data=[]dto.ParentOption}
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /api/members/search-info [get]
func (h *memberHandler) SearchMemberInfo(c *gin.Context) {
	var query dto.ParentSearchQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: err.Error()})
		return
	}

	filter := domain.MemberFilter{
		Name:   &query.Query,
		Gender: &query.Gender,
	}

	members, _, err := h.memberUseCase.SearchMembers(c.Request.Context(), filter, nil, query.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	var options []dto.ParentOption
	for _, member := range members {
		options = append(options, dto.ParentOption{
			MemberID:    member.MemberID,
			ArabicName:  member.ArabicName,
			EnglishName: member.EnglishName,
			Picture:     member.Picture,
			Gender:      member.Gender,
		})
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: options})
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

	userID := middleware.GetUserID(c)
	if err := h.memberUseCase.DeletePicture(c.Request.Context(), memberID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: "picture deleted"})
}

func (h *memberHandler) GetPicture(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("member_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "invalid member_id"})
		return
	}

	imageData, contentType, err := h.memberUseCase.GetPicture(c.Request.Context(), memberID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Response{Success: false, Error: err.Error()})
		return
	}

	c.Header("Content-Type", contentType)
	// c.Header("Cache-Control", "public, max-age=86400") // Cache for 1 day
	c.Data(http.StatusOK, contentType, imageData)
}
