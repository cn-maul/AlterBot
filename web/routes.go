package web

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/cn-maul/AlterBot/monitor"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *WebServer) setupRoutes() {
	s.engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	s.engine.GET("/api/health", s.healthCheck)
	s.engine.GET("/api/groups", s.listGroups)
	s.engine.GET("/api/settings/notifications", s.getNotificationSettings)
	s.engine.PUT("/api/settings/notifications", s.updateNotificationSettings)

	api := s.engine.Group("/api/v1/monitors")
	{
		api.GET("/", s.listMonitors)
		api.GET("/:name", s.getMonitor)
		api.POST("/", s.addMonitor)
		api.PUT("/:name", s.updateMonitor)
		api.DELETE("/:name", s.removeMonitor)
		api.POST("/:name/start", s.startMonitor)
		api.POST("/:name/stop", s.stopMonitor)
		api.GET("/:name/updates", s.getUpdates)
		api.GET("/:name/config", s.getMonitorConfig)
		api.PUT("/:name/mark-all-notified", s.markAllNotified)
	}

	// 智能扫描（在 api 组之外，避免 :name 通配符冲突）
	s.engine.POST("/api/v1/monitors/preview", s.previewScan)
	s.engine.POST("/api/v1/monitors/smart-create", s.smartCreate)

	if s.frontendFS != nil {
		assets, err := fs.Sub(s.frontendFS, "assets")
		if err == nil {
			s.engine.StaticFS("/assets", http.FS(assets))
		}
		s.engine.NoRoute(func(c *gin.Context) {
			indexHTML, err := fs.ReadFile(s.frontendFS, "index.html")
			if err != nil {
				c.String(http.StatusNotFound, "not found")
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
		})
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

func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{Code: 0, Message: "success", Data: data}
}

func NewErrorResponse(code int, message string) APIResponse {
	return APIResponse{Code: code, Message: message}
}