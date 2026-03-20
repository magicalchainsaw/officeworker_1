package gin

import (
	"officeworker/internal/handler"
	"officeworker/internal/pkg/middleware"
)

func (rg *RouterGroup) SetupAuthRoutes(authHandler *handler.AuthHandler, authMiddleware *middleware.AuthMiddleware) {
	v1 := rg.engine.Group("/api/v1/auth")
	{
		v1.POST("/register", authHandler.Register)
		v1.POST("/login", authHandler.Login)
		v1.POST("/refresh", authHandler.Refresh)
		v1.POST("/logout", authMiddleware.RequireAuth(), authHandler.Logout)
		v1.GET("/me", authMiddleware.RequireAuth(), authHandler.GetUserInfo)
	}
}
