package gin

import (
	"officeworker/internal/handler"
	"officeworker/internal/pkg/middleware"
)

func (rg *RouterGroup) SetupSessionRoutes(sessionHandler *handler.SessionHandler, authMiddleware *middleware.AuthMiddleware) {
	v1 := rg.engine.Group("/api/v1/sessions")
	v1.Use(authMiddleware.RequireAuth())
	{
		v1.POST("", sessionHandler.Create)
		v1.GET("", sessionHandler.List)
		v1.GET("/active", sessionHandler.ListActive)
		v1.GET("/:id", sessionHandler.Get)
		v1.POST("/:id", sessionHandler.SendMessage)
		v1.POST("/:id/messages", sessionHandler.SendMessage)
		v1.PUT("/:id", sessionHandler.Update)
		v1.POST("/:id/activate", sessionHandler.Activate)
		v1.POST("/:id/deactivate", sessionHandler.Deactivate)
		v1.DELETE("/:id", sessionHandler.Delete)
	}
}
