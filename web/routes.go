package web

import (
	"github.com/cn-maul/AlterBot/monitor"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *WebServer) setupRoutes() {
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
