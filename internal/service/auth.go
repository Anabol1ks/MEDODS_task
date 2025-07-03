package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"medods-auth/internal/auth/token"
	"medods-auth/internal/config"
	"medods-auth/internal/models"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		DB:  db,
		Cfg: cfg,
	}
}

func (s *AuthService) CreateTokenPair(userID, userAgent, ip string) (string, string, error) {
	access, err := token.GenerateAccessToken(userID, s.Cfg.TokenTTL)
	if err != nil {
		return "", "", err
	}
	refresh, hash, err := token.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return "", "", err
	}

	session := models.Session{
		UserID:           userUUID,
		RefreshTokenHash: hash,
		UserAgent:        userAgent,
		IP:               ip,
		ExpiresAt:        time.Now().Add(s.Cfg.RefreshTTL),
	}
	if err := s.DB.Create(&session).Error; err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) RefreshTokenPair(cfg *config.Config, accessToken, refreshToken, userAgent, ip string) (*RefreshTokenResult, error) {
	userID, err := token.ParseAccessToken(accessToken)
	if err != nil || userID == "" {
		return nil, errors.New("invalid access token")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user_id in access token")
	}

	var session models.Session
	if err := s.DB.Where("user_id = ?", userUUID).Order("created_at desc").First(&session).Error; err != nil {
		return nil, errors.New("session not found")
	}

	if err := token.ValidateRefreshToken(refreshToken, session.RefreshTokenHash, session.ExpiresAt); err != nil {
		s.DB.Delete(&session)
		return nil, errors.New("invalid or expired refresh token; session revoked")
	}

	if session.UserAgent != userAgent {
		s.DB.Delete(&session)
		return nil, errors.New("user-agent mismatch; session revoked")
	}

	if session.IP != ip {
		go sendWebhook(cfg.WebHookUrl, userID, session.IP, ip, userAgent)
	}

	s.DB.Delete(&session)

	access, refresh, err := s.CreateTokenPair(userID, userAgent, ip)
	if err != nil {
		return nil, errors.New("failed to create new token pair")
	}

	return &RefreshTokenResult{AccessToken: access, RefreshToken: refresh}, nil
}

func sendWebhook(url, userID, oldIP, newIP, userAgent string) {
	body := map[string]string{
		"user_id":    userID,
		"old_ip":     oldIP,
		"new_ip":     newIP,
		"user_agent": userAgent,
		"event":      "refresh_from_new_ip",
	}
	b, _ := json.Marshal(body)
	http.Post(url, "application/json", bytes.NewBuffer(b))
}
