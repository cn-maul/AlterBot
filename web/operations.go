package web

import (
	"github.com/cn-maul/AlterBot/config"
	"github.com/cn-maul/AlterBot/monitor"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *WebServer) addMonitor(c *gin.Context) {
	var siteConfig config.SiteConfig
	if err := c.ShouldBindJSON(&siteConfig); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(400, "invalid request body"))
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	// 检查是否已存在
	if _, exists := s.monitors[siteConfig.Name]; exists {
		c.JSON(http.StatusConflict, NewErrorResponse(409, "monitor already exists"))
		return
	}

	// 创建并启动监控器
	stopCh := make(chan struct{})
	go monitor.Start(&siteConfig, stopCh)

	// 保存上下文
	s.monitors[siteConfig.Name] = &monitorContext{
		config: &siteConfig,
		stopCh: stopCh,
	}

	c.JSON(http.StatusCreated, NewSuccessResponse(siteConfig.Name))
}

func (s *WebServer) removeMonitor(c *gin.Context) {
	name := c.Param("name")
	s.lock.Lock()
	defer s.lock.Unlock()

	if ctx, exists := s.monitors[name]; exists {
		close(ctx.stopCh)
		delete(s.monitors, name)
		c.JSON(http.StatusOK, NewSuccessResponse(nil))
		return
	}

	c.JSON(http.StatusNotFound, NewErrorResponse(404, "monitor not found"))
}

func (s *WebServer) startMonitor(c *gin.Context) {
	name := c.Param("name")
	s.lock.Lock()
	defer s.lock.Unlock()

	ctx, exists := s.monitors[name]
	if !exists {
		c.JSON(http.StatusNotFound, NewErrorResponse(404, "monitor not found"))
		return
	}

	if monitor.GetMonitor(name).GetStatus().IsRunning {
		c.JSON(http.StatusOK, NewErrorResponse(0, "monitor is already running"))
		return
	}

	// 重新创建停止通道
	ctx.stopCh = make(chan struct{})
	go monitor.Start(ctx.config, ctx.stopCh)

	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}

func (s *WebServer) stopMonitor(c *gin.Context) {
	name := c.Param("name")
	s.lock.Lock()
	defer s.lock.Unlock()

	ctx, exists := s.monitors[name]
	if !exists {
		c.JSON(http.StatusNotFound, NewErrorResponse(404, "monitor not found"))
		return
	}

	if !monitor.GetMonitor(name).GetStatus().IsRunning {
		c.JSON(http.StatusOK, NewErrorResponse(0, "monitor is already stopped"))
		return
	}

	close(ctx.stopCh)
	ctx.stopCh = make(chan struct{}) // 创建新的通道以备重新启动

	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}
