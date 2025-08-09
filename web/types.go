package web

import (
	"github.com/cn-maul/AlterBot/config"
	"github.com/gin-gonic/gin"
	"sync"
)

// WebServer 核心服务器结构
type WebServer struct {
	engine   *gin.Engine
	monitors map[string]*monitorContext
	lock     sync.RWMutex
}

// monitorContext 监控器上下文，包含配置和控制通道
type monitorContext struct {
	config *config.SiteConfig
	stopCh chan struct{}
}

// APIResponse 标准API响应格式
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewWebServer 创建WebServer实例
func NewWebServer() *WebServer {
	return &WebServer{
		engine:   gin.Default(),
		monitors: make(map[string]*monitorContext),
	}
}
