package store

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"project1/internal/middleware"
	"project1/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) MerchantGetStore(c *gin.Context) {
	item, err := h.service.GetMerchantStore(c.Request.Context(), merchantUserID(c))
	if err != nil {
		handleStoreError(c, err)
		return
	}
	response.Success(c, item)
}

type merchantStoreUpdateRequest struct {
	Name        *string      `json:"name"`
	Description *string      `json:"description"`
	Phone       *string      `json:"phone"`
	Address     *string      `json:"address"`
	Status      *StoreStatus `json:"status"`
}

func (h *Handler) MerchantUpdateStore(c *gin.Context) {
	var req merchantStoreUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Status != nil && !req.Status.IsValid() {
		response.Error(c, http.StatusBadRequest, "invalid store status")
		return
	}
	item, err := h.service.UpdateMerchantStore(c.Request.Context(), merchantUserID(c), StoreUpdateInput{
		Name:        req.Name,
		Description: req.Description,
		Phone:       req.Phone,
		Address:     req.Address,
		Status:      req.Status,
	})
	if err != nil {
		handleStoreError(c, err)
		return
	}
	response.Success(c, item)
}

type categoryRequest struct {
	Name      *string `json:"name"`
	SortOrder *int    `json:"sort_order"`
}

func (h *Handler) MerchantListCategories(c *gin.Context) {
	items, err := h.service.ListMerchantCategories(c.Request.Context(), merchantUserID(c))
	if err != nil {
		handleStoreError(c, err)
		return
	}
	response.Success(c, items)
}

func (h *Handler) MerchantCreateCategory(c *gin.Context) {
	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Name == nil || strings.TrimSpace(*req.Name) == "" {
		response.Error(c, http.StatusBadRequest, "name is required")
		return
	}
	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	item, err := h.service.CreateMerchantCategory(c.Request.Context(), merchantUserID(c), CategoryCreateInput{
		Name:      strings.TrimSpace(*req.Name),
		SortOrder: sortOrder,
	})
	if err != nil {
		handleStoreError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *Handler) MerchantUpdateCategory(c *gin.Context) {
	categoryID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Name == nil && req.SortOrder == nil {
		response.Error(c, http.StatusBadRequest, "at least one field is required")
		return
	}
	input := CategoryUpdateInput{Name: req.Name, SortOrder: req.SortOrder}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		input.Name = &name
	}
	item, err := h.service.UpdateMerchantCategory(c.Request.Context(), merchantUserID(c), categoryID, input)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *Handler) MerchantDeleteCategory(c *gin.Context) {
	categoryID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.DeleteMerchantCategory(c.Request.Context(), merchantUserID(c), categoryID); err != nil {
		handleStoreError(c, err)
		return
	}
	response.Success(c, gin.H{})
}

type dishCreateRequest struct {
	CategoryID  *int64      `json:"category_id"`
	Name        *string     `json:"name"`
	Description *string     `json:"description"`
	Price       *float64    `json:"price"`
	Stock       *int        `json:"stock"`
	Status      *DishStatus `json:"status"`
	ImageURL    *string     `json:"image_url"`
}

func (h *Handler) MerchantListDishes(c *gin.Context) {
	filter, err := parseDishFilter(c, true)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.service.ListMerchantDishes(c.Request.Context(), merchantUserID(c), filter)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	response.Success(c, items)
}

func (h *Handler) MerchantCreateDish(c *gin.Context) {
	var req dishCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.CategoryID == nil || req.Name == nil || req.Price == nil {
		response.Error(c, http.StatusBadRequest, "category_id, name and price are required")
		return
	}
	if *req.Price < 0 || (req.Stock != nil && *req.Stock < 0) {
		response.Error(c, http.StatusBadRequest, "price and stock must be >= 0")
		return
	}
	status := DishStatusOnSale
	if req.Status != nil {
		status = *req.Status
	}
	if !status.IsValid() {
		response.Error(c, http.StatusBadRequest, "invalid dish status")
		return
	}
	stock := 0
	if req.Stock != nil {
		stock = *req.Stock
	}
	description := ""
	if req.Description != nil {
		description = strings.TrimSpace(*req.Description)
	}
	imageURL := ""
	if req.ImageURL != nil {
		imageURL = strings.TrimSpace(*req.ImageURL)
	}
	item, err := h.service.CreateMerchantDish(c.Request.Context(), merchantUserID(c), DishCreateInput{
		CategoryID:  *req.CategoryID,
		Name:        strings.TrimSpace(*req.Name),
		Description: description,
		Price:       *req.Price,
		Stock:       stock,
		Status:      status,
		ImageURL:    imageURL,
	})
	if err != nil {
		handleStoreError(c, err)
		return
	}
	response.Success(c, item)
}

type dishUpdateRequest struct {
	CategoryID  *int64   `json:"category_id"`
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Stock       *int     `json:"stock"`
	ImageURL    *string  `json:"image_url"`
}

func (h *Handler) MerchantUpdateDish(c *gin.Context) {
	dishID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req dishUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.CategoryID == nil && req.Name == nil && req.Description == nil && req.Price == nil && req.Stock == nil && req.ImageURL == nil {
		response.Error(c, http.StatusBadRequest, "at least one field is required")
		return
	}
	if req.Price != nil && *req.Price < 0 {
		response.Error(c, http.StatusBadRequest, "price must be >= 0")
		return
	}
	if req.Stock != nil && *req.Stock < 0 {
		response.Error(c, http.StatusBadRequest, "stock must be >= 0")
		return
	}
	input := DishUpdateInput{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageURL,
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		input.Name = &name
	}
	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		input.Description = &description
	}
	if req.ImageURL != nil {
		imageURL := strings.TrimSpace(*req.ImageURL)
		input.ImageURL = &imageURL
	}
	item, err := h.service.UpdateMerchantDish(c.Request.Context(), merchantUserID(c), dishID, input)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	response.Success(c, item)
}

type dishStatusRequest struct {
	Status DishStatus `json:"status" binding:"required"`
}

func (h *Handler) MerchantUpdateDishStatus(c *gin.Context) {
	dishID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req dishStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if !req.Status.IsValid() {
		response.Error(c, http.StatusBadRequest, "invalid dish status")
		return
	}
	item, err := h.service.UpdateDishStatus(c.Request.Context(), merchantUserID(c), dishID, req.Status)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *Handler) PublicListStores(c *gin.Context) {
	items, err := h.service.ListPublicStores(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, items)
}

func (h *Handler) PublicGetStore(c *gin.Context) {
	storeID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	item, err := h.service.GetPublicStore(c.Request.Context(), storeID)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *Handler) PublicListCategories(c *gin.Context) {
	storeID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	items, err := h.service.ListPublicCategories(c.Request.Context(), storeID)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	response.Success(c, items)
}

func (h *Handler) PublicListDishes(c *gin.Context) {
	storeID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	filter, err := parseDishFilter(c, false)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.service.ListPublicDishes(c.Request.Context(), storeID, filter)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	response.Success(c, items)
}

func (h *Handler) PublicListReviews(c *gin.Context) {
	storeID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	items, err := h.service.ListPublicReviews(c.Request.Context(), storeID)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	response.Success(c, items)
}

func parseDishFilter(c *gin.Context, allowStatus bool) (DishQueryFilter, error) {
	filter := DishQueryFilter{}
	if raw := strings.TrimSpace(c.Query("category_id")); raw != "" {
		categoryID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return DishQueryFilter{}, errors.New("invalid category_id")
		}
		filter.CategoryID = &categoryID
	}
	filter.Keyword = strings.TrimSpace(c.Query("keyword"))
	if allowStatus {
		rawStatus := strings.TrimSpace(c.Query("status"))
		if rawStatus != "" {
			status := DishStatus(rawStatus)
			if !status.IsValid() {
				return DishQueryFilter{}, errors.New("invalid dish status")
			}
			filter.Status = &status
		}
	}
	return filter, nil
}

func parseIDParam(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func merchantUserID(c *gin.Context) int64 {
	claims, _ := middleware.ClaimsFromContext(c)
	return claims.UserID
}

func handleStoreError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(c, http.StatusNotFound, "resource not found")
	case errors.Is(err, ErrCategoryInUse):
		response.Error(c, http.StatusBadRequest, "category has dishes, delete dishes first")
	case errors.Is(err, ErrInvalidRelation):
		response.Error(c, http.StatusBadRequest, "invalid category or relation")
	case errors.Is(err, ErrDuplicateName):
		response.Error(c, http.StatusBadRequest, "duplicate name in same store")
	default:
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
}
