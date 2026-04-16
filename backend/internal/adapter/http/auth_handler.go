package http

import (
	"errors"
	"net/http"

	"acc-dp/backend/internal/domain"
	"acc-dp/backend/internal/service/auth"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *auth.Service
}

func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type registerRequest struct {
	Email       string `json:"email" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

type loginRequest struct {
	Email      string `json:"email" binding:"required"`
	Password   string `json:"password" binding:"required"`
	MachineID  string `json:"machine_id" binding:"required"`
	DeviceName string `json:"device_name"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	MachineID    string `json:"machine_id" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	MachineID    string `json:"machine_id" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	result, err := h.authService.Register(c.Request.Context(), auth.RegisterInput{
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Password:    req.Password,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	result, err := h.authService.Login(c.Request.Context(), auth.LoginInput{
		Email:      req.Email,
		Password:   req.Password,
		MachineID:  req.MachineID,
		DeviceName: req.DeviceName,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	result, err := h.authService.Refresh(c.Request.Context(), auth.RefreshInput{
		RefreshToken: req.RefreshToken,
		MachineID:    req.MachineID,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	err := h.authService.Logout(c.Request.Context(), auth.LogoutInput{
		RefreshToken: req.RefreshToken,
		MachineID:    req.MachineID,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "logged_out"})
}

func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrUserAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidRefreshToken):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrMachineMismatch):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
