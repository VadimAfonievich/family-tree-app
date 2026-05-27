package handler

import (
	"net/http"

	"family-tree-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userService *service.UserService
}

func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

type authRequest struct {
	InitData string `json:"init_data" binding:"required"`
}

type authResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Telegram(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.userService.AuthByTelegram(c.Request.Context(), req.InitData)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid telegram data"})
		return
	}

	c.JSON(http.StatusOK, authResponse{Token: token})
}
