package web

import (
	"github.com/cn-maul/AlterBot/monitor"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func (s *WebServer) setupRoutes() {
	// 添加CORS配置
	s.engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // 替换为你的前端地址
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := s.engine.Group("/api/v1/monitors")
	{
		api.GET("/", s.listMonitors)
		api.GET("/:name", s.getMonitor)
		api.POST("/", s.addMonitor)
		api.DELETE("/:name", s.removeMonitor)
		api.POST("/:name/start", s.startMonitor)
		api.POST("/:name/stop", s.stopMonitor)
	}
}

func (s *WebServer) listMonitors(c *gin.Context) {
	monitors := monitor.GetAllMonitors()
	c.JSON(http.StatusOK, NewSuccessResponse(monitors))
}

func (s *WebServer) getMonitor(c *gin.Context) {
	name := c.Param("name")
	m := monitor.GetMonitor(name)
	if m == nil {
		c.JSON(http.StatusNotFound, NewErrorResponse(404, "monitor not found"))
		return
	}
	c.JSON(http.StatusOK, NewSuccessResponse(m.GetStatus()))
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) APIResponse {
	return APIResponse{
		Code:    code,
		Message: message,
	}
}
