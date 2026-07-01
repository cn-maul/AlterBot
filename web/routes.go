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
	// 添加CORS配置
	s.engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 健康检查
	s.engine.GET("/api/health", s.healthCheck)

	// 分组列表
	s.engine.GET("/api/groups", s.listGroups)

	// 通知设置
	s.engine.GET("/api/settings/notifications", s.getNotificationSettings)
	s.engine.PUT("/api/settings/notifications", s.updateNotificationSettings)

	// 监控器管理 API
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
	}

	// 嵌入的前端静态文件（生产模式自动启用）
	if s.frontendFS != nil {
		// 提取 assets 子目录作为文件服务器
		assets, err := fs.Sub(s.frontendFS, "assets")
		if err == nil {
			s.engine.StaticFS("/assets", http.FS(assets))
		}
		// SPA 回退：所有未匹配的路由返回 index.html
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
