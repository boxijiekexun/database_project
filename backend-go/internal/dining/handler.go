package dining

import (
	"net/http"
	"strconv"
	"strings"

	"project1/internal/response"
	"project1/internal/store"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListAreas(c *gin.Context) {
	items, err := h.service.ListAreas(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, items)
}

func (h *Handler) ListCanteens(c *gin.Context) {
	areaID, ok := parseDiningID(c, "id")
	if !ok {
		return
	}
	items, err := h.service.ListCanteens(c.Request.Context(), areaID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, items)
}

func (h *Handler) GetCanteen(c *gin.Context) {
	canteenID, ok := parseDiningID(c, "id")
	if !ok {
		return
	}
	item, err := h.service.GetCanteen(c.Request.Context(), canteenID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *Handler) ListStores(c *gin.Context) {
	canteenID, ok := parseDiningID(c, "id")
	if !ok {
		return
	}
	items, err := h.service.ListStores(c.Request.Context(), canteenID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, items)
}

func (h *Handler) ListDishes(c *gin.Context) {
	canteenID, ok := parseDiningID(c, "id")
	if !ok {
		return
	}
	filter, err := parseCanteenDishFilter(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.service.ListDishes(c.Request.Context(), canteenID, filter)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, items)
}

func parseCanteenDishFilter(c *gin.Context) (store.DishQueryFilter, error) {
	var filter store.DishQueryFilter
	if value := strings.TrimSpace(c.Query("keyword")); value != "" {
		filter.Keyword = value
	}
	if value := strings.TrimSpace(c.Query("category_id")); value != "" {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return store.DishQueryFilter{}, err
		}
		filter.CategoryID = &id
	}
	return filter, nil
}

func parseDiningID(c *gin.Context, key string) (int64, bool) {
	value, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil || value <= 0 {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return value, true
}
