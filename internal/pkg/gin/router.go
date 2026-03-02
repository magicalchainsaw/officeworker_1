package gin

import (
	"github.com/gin-gonic/gin"
)

type RouterGroup struct {
	engine *gin.Engine
}

func NewRouterGroup(engine *gin.Engine) *RouterGroup {
	return &RouterGroup{engine: engine}
}

func (rg *RouterGroup) SetupRoutes() {
	v1 := rg.engine.Group("/api/v1")

	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
}
