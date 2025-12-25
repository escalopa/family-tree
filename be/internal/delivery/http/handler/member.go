package handler

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/escalopa/family-tree/internal/delivery"
	"github.com/escalopa/family-tree/internal/delivery/http/dto"
	"github.com/escalopa/family-tree/internal/delivery/http/middleware"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

type memberHandler struct {
	memberUseCase   MemberUseCase
	languageUseCase LanguageUseCase
}

func NewMemberHandler(memberUseCase MemberUseCase, languageUseCase LanguageUseCase) *memberHandler {
	return &memberHandler{
		memberUseCase:   memberUseCase,
		languageUseCase: languageUseCase,
	}
}

func (h *memberHandler) validateNames(c *gin.Context, names map[string]string) error {
	activeLanguages, err := h.languageUseCase.List(c.Request.Context(), true)
	if err != nil {
		delivery.Error(c, err)
		return err
	}

	for _, lang := range activeLanguages {
		name, exists := names[lang.LanguageCode]
		if !exists || name == "" {
			validationErr := domain.NewValidationError("error.validation.names_required", map[string]string{
				"language": lang.LanguageName,
				"code":     lang.LanguageCode,
			})
			delivery.Error(c, validationErr)
			return validationErr
		}
	}

	return nil
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
func (h *memberHandler) Create(c *gin.Context) {
	var req dto.CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		delivery.Error(c, err)
		return
	}

	if err := h.validateNames(c, req.Names); err != nil {
		return
	}
	// Ensure nicknames is never nil, use empty array instead
	nicknames := req.Nicknames
	if nicknames == nil {
		nicknames = []string{}
	}

	member := &domain.Member{
		Names:       req.Names,
		Gender:      req.Gender,
		DateOfBirth: req.DateOfBirth.ToTimePtr(),
		DateOfDeath: req.DateOfDeath.ToTimePtr(),
		FatherID:    req.FatherID,
		MotherID:    req.MotherID,
		Nicknames:   nicknames,
		Profession:  req.Profession,
	}

	userID := middleware.GetUserID(c)
	if err := h.memberUseCase.Create(c.Request.Context(), member, userID); err != nil {
		delivery.Error(c, err)
		return
	}

	response := gin.H{
		"member_id": member.MemberID,
		"version":   member.Version,
	}

	delivery.SuccessWithData(c, response)
}

func (h *memberHandler) Update(c *gin.Context) {
	var uri dto.MemberIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	var req dto.UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		delivery.Error(c, err)
		return
	}

	if err := h.validateNames(c, req.Names); err != nil {
		return
	}

	existingMember, err := h.memberUseCase.Get(c.Request.Context(), uri.MemberID)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	nicknames := req.Nicknames
	if nicknames == nil {
		nicknames = []string{}
	}

	member := &domain.Member{
		MemberID:    uri.MemberID,
		Names:       req.Names,
		Gender:      req.Gender,
		Picture:     existingMember.Picture,
		DateOfBirth: req.DateOfBirth.ToTimePtr(),
		DateOfDeath: req.DateOfDeath.ToTimePtr(),
		FatherID:    req.FatherID,
		MotherID:    req.MotherID,
		Nicknames:   nicknames,
		Profession:  req.Profession,
	}

	userID := middleware.GetUserID(c)
	if err := h.memberUseCase.Update(c.Request.Context(), member, req.Version, userID); err != nil {
		delivery.Error(c, err)
		return
	}

	response := gin.H{"version": member.Version}

	delivery.SuccessWithData(c, response)
}

func (h *memberHandler) Delete(c *gin.Context) {
	var uri dto.MemberIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.memberUseCase.Delete(c.Request.Context(), uri.MemberID, userID); err != nil {
		delivery.Error(c, err)
		return
	}

	delivery.Success(c, "success.member.deleted", nil)
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
func (h *memberHandler) Get(c *gin.Context) {
	var uri dto.MemberIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	member, err := h.memberUseCase.Get(c.Request.Context(), uri.MemberID)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	userRole := middleware.GetUserRole(c)
	computed := h.memberUseCase.Compute(c.Request.Context(), member, userRole)

	preferredLang := middleware.GetPreferredLanguage(c)
	memberID := uri.MemberID

	spousesDTO := make([]dto.SpouseInfo, len(computed.Spouses))
	for i, spouse := range computed.Spouses {
		spousesDTO[i] = dto.SpouseInfo{
			SpouseID:     spouse.SpouseID,
			MemberID:     spouse.MemberID,
			Name:         extractName(spouse.Names, preferredLang),
			Gender:       spouse.Gender,
			Picture:      spouse.Picture,
			MarriageDate: dto.FromTimePtr(spouse.MarriageDate),
			DivorceDate:  dto.FromTimePtr(spouse.DivorceDate),
			MarriedYears: dto.CalculateMarriedYears(spouse.MarriageDate, spouse.DivorceDate),
		}
	}

	var fatherInfo, motherInfo *dto.MemberInfo
	if computed.FatherID != nil {
		father, err := h.memberUseCase.Get(c.Request.Context(), *computed.FatherID)
		if err == nil {
			fatherInfo = &dto.MemberInfo{
				MemberID: father.MemberID,
				Name:     extractName(father.Names, preferredLang),
				Picture:  father.Picture,
			}
		}
	}
	if computed.MotherID != nil {
		mother, err := h.memberUseCase.Get(c.Request.Context(), *computed.MotherID)
		if err == nil {
			motherInfo = &dto.MemberInfo{
				MemberID: mother.MemberID,
				Name:     extractName(mother.Names, preferredLang),
				Picture:  mother.Picture,
			}
		}
	}

	var childrenInfo []dto.MemberInfo
	children, err := h.memberUseCase.ListChildren(c.Request.Context(), memberID)
	if err == nil {
		for _, child := range children {
			childrenInfo = append(childrenInfo, dto.MemberInfo{
				MemberID: child.MemberID,
				Name:     extractName(child.Names, preferredLang),
				Picture:  child.Picture,
			})
		}
	}

	var siblingsInfo []dto.MemberInfo
	siblings, err := h.memberUseCase.ListSiblings(c.Request.Context(), memberID)
	if err == nil {
		for _, sibling := range siblings {
			siblingsInfo = append(siblingsInfo, dto.MemberInfo{
				MemberID: sibling.MemberID,
				Name:     extractName(sibling.Names, preferredLang),
				Picture:  sibling.Picture,
			})
		}
	}

	response := dto.MemberResponse{
		MemberID:        computed.MemberID,
		Name:            extractName(computed.Names, preferredLang),
		Names:           computed.Names,
		FullName:        extractName(computed.FullNames, preferredLang),
		FullNames:       computed.FullNames,
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
		Age:             computed.Age,
		GenerationLevel: computed.GenerationLevel,
		IsMarried:       computed.IsMarried,
		Spouses:         spousesDTO,
		Children:        childrenInfo,
		Siblings:        siblingsInfo,
	}

	delivery.SuccessWithData(c, response)
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
func (h *memberHandler) List(c *gin.Context) {
	var query dto.MemberListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		delivery.Error(c, err)
		return
	}

	filter := domain.MemberFilter{
		Name:      query.Name,
		Gender:    query.Gender,
		IsMarried: query.Married,
	}

	members, nextCursor, err := h.memberUseCase.List(c.Request.Context(), filter, query.Cursor, query.Limit)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	preferredLang := middleware.GetPreferredLanguage(c)

	var membersResponse []dto.MemberListItem
	for _, m := range members {
		membersResponse = append(membersResponse, dto.MemberListItem{
			MemberID:    m.MemberID,
			Name:        extractName(m.Names, preferredLang),
			Gender:      m.Gender,
			Picture:     m.Picture,
			DateOfBirth: dto.FromTimePtr(m.DateOfBirth),
			DateOfDeath: dto.FromTimePtr(m.DateOfDeath),
			IsMarried:   m.IsMarried,
		})
	}

	response := dto.PaginatedMembersResponse{
		Members:    membersResponse,
		NextCursor: nextCursor,
	}

	delivery.SuccessWithData(c, response)
}

func (h *memberHandler) ListHistory(c *gin.Context) {
	var query dto.MemberHistoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		delivery.Error(c, err)
		return
	}

	history, nextCursor, err := h.memberUseCase.ListHistory(c.Request.Context(), query.MemberID, query.Cursor, query.Limit)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	preferredLang := middleware.GetPreferredLanguage(c)

	var historyResponse []dto.HistoryResponse
	for _, h := range history {
		historyResponse = append(historyResponse, dto.HistoryResponse{
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
		History:    historyResponse,
		NextCursor: nextCursor,
	}

	delivery.SuccessWithData(c, response)
}

func (h *memberHandler) UploadPicture(c *gin.Context) {
	var uri dto.MemberIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	file, header, err := c.Request.FormFile("picture")
	if err != nil {
		delivery.Error(c, domain.NewValidationError("error.invalid_input", map[string]string{"message": "missing picture file"}))
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		slog.Error("memberHandler.UploadPicture: read file", "error", err, "member_id", uri.MemberID)
		delivery.Error(c, domain.NewInternalError(err))
		return
	}

	userID := middleware.GetUserID(c)
	pictureURL, err := h.memberUseCase.UploadPicture(c.Request.Context(), uri.MemberID, data, header.Filename, userID)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	response := gin.H{"picture_url": pictureURL}

	delivery.SuccessWithData(c, response)
}

func (h *memberHandler) DeletePicture(c *gin.Context) {
	var uri dto.MemberIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.memberUseCase.DeletePicture(c.Request.Context(), uri.MemberID, userID); err != nil {
		delivery.Error(c, err)
		return
	}

	delivery.Success(c, "success.member.picture_deleted", nil)
}

func (h *memberHandler) GetPicture(c *gin.Context) {
	var uri dto.MemberIDUri
	if err := c.ShouldBindUri(&uri); err != nil {
		delivery.Error(c, err)
		return
	}

	imageData, contentType, err := h.memberUseCase.GetPicture(c.Request.Context(), uri.MemberID)
	if err != nil {
		delivery.Error(c, err)
		return
	}

	c.Header("Content-Type", contentType)
	// c.Header("Cache-Control", "public, max-age=86400") // Cache for 1 day
	c.Data(http.StatusOK, contentType, imageData)
}
