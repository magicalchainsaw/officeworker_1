package middleware

import (
	"errors"
	"net/http"
	"officeworker/internal/pkg/jwt"
	"officeworker/internal/pkg/redis"
	"officeworker/internal/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtMgr    *jwt.Manager
	blacklist *redis.Blacklist
}

func NewAuthMiddleware(jwtMgr *jwt.Manager, blacklist *redis.Blacklist) *AuthMiddleware {
	return &AuthMiddleware{
		jwtMgr:    jwtMgr,
		blacklist: blacklist,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			response.AbortWithError(c, http.StatusUnauthorized, err.Error())
			return
		}

		blacklisted, err := m.blacklist.Exists(tokenString)
		if err != nil {
			response.AbortWithError(c, http.StatusInternalServerError, "failed to check token blacklist")
			return
		}
		if blacklisted {
			response.AbortWithError(c, http.StatusUnauthorized, "token has been revoked")
			return
		}

		claims, err := m.jwtMgr.ParseToken(tokenString)
		if err != nil {
			response.AbortWithError(c, http.StatusUnauthorized, "invalid token")
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("access_token", tokenString)

		c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			response.AbortWithError(c, http.StatusUnauthorized, "user not authenticated")
			return
		}

		role := userRole.(string)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		response.AbortWithError(c, http.StatusForbidden, "insufficient permissions")
		c.Abort()
	}
}

func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid authorization header format")
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" {
		return "", errors.New("missing bearer token")
	}

	return token, nil
}
