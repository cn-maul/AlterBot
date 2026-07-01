package web

import (
	"io/fs"

	"github.com/gin-gonic/gin"
)

// WebServer 核心服务器结构
type WebServer struct {
	engine     *gin.Engine
	frontendFS fs.FS
}

// APIResponse 标准API响应格式
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewWebServer 创建WebServer实例
func NewWebServer(frontendFS fs.FS) *WebServer {
	return &WebServer{engine: gin.Default(), frontendFS: frontendFS}
}