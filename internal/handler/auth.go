package handler

import (
	"medods-auth/internal/config"
	"medods-auth/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	Service *service.AuthService
	Log     *zap.Logger
}

func NewAuthHandler(s *service.AuthService, l *zap.Logger) *AuthHandler {
	return &AuthHandler{
		Service: s,
		Log:     l,
	}
}

func (s *AuthHandler) Token(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		s.Log.Warn("Token endpoint called without user_id", zap.String("user_id", userID))
		return
	}

	ua := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	access, refresh, err := s.Service.CreateTokenPair(userID, ua, ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		s.Log.Error("Failed to generate tokens", zap.String("user_id", userID), zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": access, "refresh_token": refresh})
}

type RefreshRequest struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (s *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	ua := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	cfg := config.Load(s.Log)
	result, err := s.Service.RefreshTokenPair(cfg, req.AccessToken, req.RefreshToken, ua, ip)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		s.Log.Warn("Refresh failed", zap.Error(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": result.AccessToken, "refresh_token": result.RefreshToken})
}
