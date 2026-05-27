package user

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"project1/internal/response"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	service *AdminService
}

func NewAdminHandler(service *AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

type adminStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type adminRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type adminResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	filter := AdminUserFilter{
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Role:     strings.TrimSpace(c.Query("role")),
		Status:   strings.TrimSpace(c.Query("status")),
		Page:     parseIntQuery(c, "page", 1),
		PageSize: parseIntQuery(c, "page_size", 10),
	}
	result, err := h.service.ListUsers(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	userID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	item, err := h.service.GetUser(c.Request.Context(), userID)
	if err != nil {
		handleAdminUserError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *AdminHandler) UpdateStatus(c *gin.Context) {
	userID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req adminStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	claims, ok := currentAdminClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	item, err := h.service.UpdateStatus(c.Request.Context(), claims.UserID, userID, req.Status)
	if err != nil {
		handleAdminUserError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *AdminHandler) UpdateRole(c *gin.Context) {
	userID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req adminRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	claims, ok := currentAdminClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	item, err := h.service.UpdateRole(c.Request.Context(), claims.UserID, userID, req.Role)
	if err != nil {
		handleAdminUserError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *AdminHandler) ResetPassword(c *gin.Context) {
	userID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req adminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	item, newPassword, err := h.service.ResetPassword(c.Request.Context(), userID, req.NewPassword)
	if err != nil {
		handleAdminUserError(c, err)
		return
	}
	response.Success(c, gin.H{"username": item.Username, "new_password": newPassword})
}

func handleAdminUserError(c *gin.Context, err error) {
	switch err {
	case ErrAdminUserNotFound:
		response.Error(c, http.StatusNotFound, "user not found")
	case ErrAdminUserSelfModify:
		response.Error(c, http.StatusBadRequest, err.Error())
	case ErrAdminUserInvalidRole, ErrAdminUserInvalidStatus:
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
}

func parseIntQuery(c *gin.Context, key string, fallback int) int {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseIDParam(c *gin.Context, key string) (int64, bool) {
	value, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil || value <= 0 {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return value, true
}

func currentAdminClaims(c *gin.Context) (struct{ UserID int64 }, bool) {
	value, exists := c.Get("authClaims")
	if !exists {
		return struct{ UserID int64 }{}, false
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return struct{ UserID int64 }{}, false
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() || rv.Kind() != reflect.Struct {
		return struct{ UserID int64 }{}, false
	}
	field := rv.FieldByName("UserID")
	if !field.IsValid() {
		return struct{ UserID int64 }{}, false
	}
	if field.Kind() < reflect.Int || field.Kind() > reflect.Int64 {
		return struct{ UserID int64 }{}, false
	}
	return struct{ UserID int64 }{UserID: field.Int()}, true
}
