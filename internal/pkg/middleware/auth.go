package middleware

import (
	"net/http"
	"officeworker/internal/pkg/jwt"
	"officeworker/internal/pkg/redis"
	"officeworker/internal/pkg/response"

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
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.AbortWithError(c, http.StatusUnauthorized, "missing authorization header")
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			response.AbortWithError(c, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		tokenString := authHeader[7:]

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
