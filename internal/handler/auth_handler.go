package handler

import (
	"net/http"
	"officeworker/internal/pkg/jwt"
	"officeworker/internal/pkg/response"
	"officeworker/internal/service"
	"officeworker/internal/pkg/redis"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
	jwtMgr      *jwt.Manager
	blacklist   *redis.Blacklist
	jwtExpiry   time.Duration
}

func NewAuthHandler(
	authService *service.AuthService,
	jwtMgr *jwt.Manager,
	blacklist *redis.Blacklist,
	jwtExpiry time.Duration,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtMgr:      jwtMgr,
		blacklist:   blacklist,
		jwtExpiry:   jwtExpiry,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err.Error())
		return
	}

	resp, err := h.authService.Register(&req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err.Error())
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err.Error())
		return
	}

	resp, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Error(c, "missing authorization header")
		return
	}

	accessToken := authHeader[7:]
	if err := h.authService.Logout(accessToken); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "user not authenticated")
		return
	}

	userInfo, err := h.authService.GetUserInfo(userID.(uint))
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, userInfo)
}
