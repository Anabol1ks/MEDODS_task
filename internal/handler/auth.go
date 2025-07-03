package handler

import (
	"medods-auth/internal/auth/token"
	"medods-auth/internal/config"
	"medods-auth/internal/models"
	"medods-auth/internal/service"
	"net/http"
	"strings"

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

// Token godoc
//	@Summary		Получить пару токенов
//	@Description	Генерирует access и refresh токены для пользователя
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			user_id	query		string					true	"User ID (GUID)"
//	@Success		200		{object}	response.TokensResponse	"Токены успешно сгенерированы"
//	@Failure		400		{object}	response.ErrorResponse	"Неверный запрос"
//	@Failure		500		{object}	response.ErrorResponse	"Внутренняя ошибка сервера"
//	@Router			/auth/token [post]
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

// Refresh godoc
//	@Summary		Обновить пару токенов
//	@Description	Обновляет access и refresh токены по действующей паре
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RefreshRequest			true	"Текущая пара токенов"
//	@Success		200		{object}	response.TokensResponse	"Токены успешно обновлены"
//	@Failure		400		{object}	response.ErrorResponse	"Неверный запрос"
//	@Failure		401		{object}	response.ErrorResponse	"Неавторизованный доступ"
//	@Router			/auth/refresh [post]
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

// Me godoc
//	@Summary		Получить GUID текущего пользователя
//	@Description	Возвращает user_id из access токена
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.UserResponse	"Успешно получен user_id"
//	@Failure		401	{object}	response.ErrorResponse	"Неавторизованный доступ"
//	@Router			/me [get]
func (s *AuthHandler) Me(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
		return
	}
	var tokenStr string
	if strings.HasPrefix(accessToken, "Bearer ") {
		tokenStr = accessToken[7:]
	} else {
		tokenStr = accessToken
	}
	userID, err := token.ParseAccessToken(tokenStr)
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired access token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}

// Logout godoc
//	@Summary		Деавторизация пользователя
//	@Description	Деавторизация по access токену (удаление всех сессий)
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.LogoutResponse	"Успешная деавторизация"
//	@Failure		401	{object}	response.ErrorResponse	"Неавторизованный доступ"
//	@Failure		500	{object}	response.ErrorResponse	"Внутренняя ошибка сервера"
//	@Router			/auth/logout [post]
func (s *AuthHandler) Logout(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
		return
	}
	var tokenStr string
	if len(accessToken) > 7 && accessToken[:7] == "Bearer " {
		tokenStr = accessToken[7:]
	} else {
		tokenStr = accessToken
	}
	userID, err := token.ParseAccessToken(tokenStr)
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired access token"})
		return
	}
	if err := s.Service.DB.Where("user_id = ?", userID).Delete(&models.Session{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}
