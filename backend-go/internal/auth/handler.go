package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"project1/internal/response"
	"project1/internal/user"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	repo         *user.Repository
	tokenManager *TokenManager
	captchaStore *CaptchaStore
}

func NewHandler(repo *user.Repository, tokenManager *TokenManager, captchaStore *CaptchaStore) *Handler {
	return &Handler{repo: repo, tokenManager: tokenManager, captchaStore: captchaStore}
}

type loginRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

type registerRequest struct {
	Username    string    `json:"username" binding:"required"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Password    string    `json:"password" binding:"required"`
	Role        user.Role `json:"role" binding:"required"`
	CaptchaID   string    `json:"captcha_id"`
	CaptchaCode string    `json:"captcha_code"`
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	identifier := firstNonEmpty(req.Username, req.Email, req.Phone)
	if identifier == "" {
		response.Error(c, http.StatusBadRequest, "username, email or phone is required")
		return
	}

	if err := h.verifyCaptcha(req.CaptchaID, req.CaptchaCode); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.repo.FindByLogin(c.Request.Context(), identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.Error(c, http.StatusUnauthorized, "invalid credentials")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(item.PasswordHash), []byte(req.Password)); err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := h.tokenManager.Generate(item.ID, item.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"token": token, "user": publicUser(item)})
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if !req.Role.IsValid() || !req.Role.CanRegister() {
		response.Error(c, http.StatusBadRequest, "only STUDENT or TEACHER can register")
		return
	}

	if err := h.verifyCaptcha(req.CaptchaID, req.CaptchaCode); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "hash password failed")
		return
	}

	item, err := h.repo.Create(c.Request.Context(), user.CreateInput{
		Username:     req.Username,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hash),
		Role:         req.Role,
		Status:       "ACTIVE",
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.tokenManager.Generate(item.ID, item.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"token": token, "user": publicUser(item)})
}

func (h *Handler) Captcha(c *gin.Context) {
	captcha, err := h.captchaStore.Generate()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"captcha_id": captcha.ID, "captcha_code": captcha.Code})
}

func (h *Handler) verifyCaptcha(id, code string) error {
	if id == "" && code == "" {
		return nil
	}
	if id == "" || code == "" {
		return errors.New("captcha_id and captcha_code are required together")
	}
	if !h.captchaStore.Verify(id, code) {
		return errors.New("invalid captcha")
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func publicUser(item user.User) gin.H {
	return gin.H{
		"id":       item.ID,
		"username": item.Username,
		"email":    item.Email,
		"phone":    item.Phone,
		"role":     item.Role,
		"status":   item.Status,
	}
}
