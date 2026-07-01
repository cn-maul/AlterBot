package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cn-maul/AlterBot/database"
	"github.com/cn-maul/AlterBot/monitor"
	"github.com/cn-maul/AlterBot/notify"
	"github.com/gin-gonic/gin"
)

// addMonitorRequest 创建监控器的请求体
type addMonitorRequest struct {
	Name          string              `json:"name" binding:"required"`
	URL           string              `json:"url" binding:"required"`
	Container     string              `json:"container" binding:"required"`
	Item          string              `json:"item"`
	Group         string              `json:"group"`
	CheckInterval int                 `json:"check_interval"`
	IsActive      bool                `json:"is_active"`
	Fields        []fieldRequest      `json:"fields"`
}

type fieldRequest struct {
	Name      string `json:"name" binding:"required"`
	Selector  string `json:"selector" binding:"required"`
	Type      string `json:"type"`
	Attr      string `json:"attr"`
	Transform string `json:"transform"`
}

// dbSiteFromRequest 从请求体构建 database.Site
func dbSiteFromRequest(req *addMonitorRequest) *database.Site {
	group := req.Group
	if group == "" {
		group = "默认"
	}
	site := &database.Site{
		Name:          req.Name,
		URL:           req.URL,
		Container:     req.Container,
		Item:          req.Item,
		GroupName:     group,
		CheckInterval: req.CheckInterval,
		IsActive:      req.IsActive,
	}
	for _, f := range req.Fields {
		ft := f.Type
		if ft == "" {
			ft = "text"
		}
		site.Fields = append(site.Fields, database.SiteField{
			Name:      f.Name,
			Selector:  f.Selector,
			Type:      ft,
			Attr:      f.Attr,
			Transform: f.Transform,
		})
	}
	return site
}

// startMonitorGoroutine 启动监控器 goroutine 并注册到全局
func startMonitorGoroutine(site *database.Site) {
	go monitor.Start(site)
}

func (s *WebServer) addMonitor(c *gin.Context) {
	var req addMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(400, "invalid request body: "+err.Error()))
		return
	}

	site := dbSiteFromRequest(&req)

	if err := database.GetDB().Create(site).Error; err != nil {
		c.JSON(http.StatusConflict, NewErrorResponse(409, "monitor already exists: "+err.Error()))
		return
	}

	if site.IsActive {
		startMonitorGoroutine(site)
	}

	log.Printf("[Web] 新增监控器: %s", site.Name)
	c.JSON(http.StatusCreated, NewSuccessResponse(map[string]interface{}{
		"id":   site.ID,
		"name": site.Name,
	}))
}

func (s *WebServer) removeMonitor(c *gin.Context) {
	name := c.Param("name")

	// 如果正在运行则停止
	if monitor.Exists(name) {
		monitor.StopMonitor(name)
	}

	// 从数据库删除
	result := database.GetDB().Where("name = ?", name).Delete(&database.Site{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, NewErrorResponse(404, "monitor not found"))
		return
	}

	log.Printf("[Web] 删除监控器: %s", name)
	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}

func (s *WebServer) startMonitor(c *gin.Context) {
	name := c.Param("name")

	if monitor.Exists(name) {
		if monitor.GetMonitor(name).GetStatus().IsRunning {
			c.JSON(http.StatusOK, NewErrorResponse(0, "monitor is already running"))
			return
		}
	}

	// 从数据库读取配置
	var site database.Site
	if err := database.GetDB().Preload("Fields").Where("name = ?", name).First(&site).Error; err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse(404, "monitor not found"))
		return
	}

	site.IsActive = true
	database.GetDB().Save(&site)

	startMonitorGoroutine(&site)
	log.Printf("[Web] 启动监控器: %s", name)
	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}

func (s *WebServer) stopMonitor(c *gin.Context) {
	name := c.Param("name")

	if !monitor.Exists(name) {
		c.JSON(http.StatusNotFound, NewErrorResponse(404, "monitor not found"))
		return
	}

	if !monitor.GetMonitor(name).GetStatus().IsRunning {
		c.JSON(http.StatusOK, NewErrorResponse(0, "monitor is already stopped"))
		return
	}

	monitor.StopMonitor(name)

	// 更新数据库标记
	database.GetDB().Model(&database.Site{}).Where("name = ?", name).Update("is_active", false)

	log.Printf("[Web] 停止监控器: %s", name)
	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}

func (s *WebServer) updateMonitor(c *gin.Context) {
	name := c.Param("name")

	var req addMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(400, "invalid request body: "+err.Error()))
		return
	}

	// 从数据库查找
	var site database.Site
	if err := database.GetDB().Where("name = ?", name).First(&site).Error; err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse(404, "monitor not found"))
		return
	}

	// 停止旧实例
	if monitor.Exists(name) {
		monitor.StopMonitor(name)
	}

	// 更新数据库
	group := req.Group
	if group == "" {
		group = "默认"
	}
	database.GetDB().Model(&site).Updates(map[string]interface{}{
		"Name":          req.Name,
		"URL":           req.URL,
		"Container":     req.Container,
		"Item":          req.Item,
		"GroupName":     group,
		"CheckInterval": req.CheckInterval,
		"IsActive":      req.IsActive,
	})

	// 重设字段
	if req.Name != name {
		database.GetDB().Model(&site).Update("name", req.Name)
	}
	database.GetDB().Where("site_id = ?", site.ID).Delete(&database.SiteField{})
	for _, f := range req.Fields {
		ft := f.Type
		if ft == "" {
			ft = "text"
		}
		database.GetDB().Create(&database.SiteField{
			SiteID:    site.ID,
			Name:      f.Name,
			Selector:  f.Selector,
			Type:      ft,
			Attr:      f.Attr,
			Transform: f.Transform,
		})
	}

	// 如果要求启动则启动
	if req.IsActive {
		var updatedSite database.Site
		database.GetDB().Preload("Fields").First(&updatedSite, site.ID)
		startMonitorGoroutine(&updatedSite)
	}

	log.Printf("[Web] 更新监控器: %s -> %s", name, req.Name)
	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}

func (s *WebServer) getUpdates(c *gin.Context) {
	name := c.Param("name")

	var site database.Site
	if err := database.GetDB().Where("name = ?", name).First(&site).Error; err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse(404, "monitor not found"))
		return
	}

	page := 1
	pageSize := 20

	var total int64
	database.GetDB().Model(&database.UpdateRecord{}).
		Where("site_id = ?", site.ID).
		Count(&total)

	var records []database.UpdateRecord
	database.GetDB().Where("site_id = ?", site.ID).
		Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records)

	c.JSON(http.StatusOK, NewSuccessResponse(map[string]interface{}{
		"total":   total,
		"page":    page,
		"size":    pageSize,
		"records": records,
	}))
}

func (s *WebServer) getMonitorConfig(c *gin.Context) {
	name := c.Param("name")

	var site database.Site
	if err := database.GetDB().Preload("Fields").Where("name = ?", name).First(&site).Error; err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse(404, "monitor not found"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(site))
}

func (s *WebServer) listGroups(c *gin.Context) {
	var groups []string
	database.GetDB().Model(&database.Site{}).
		Select("DISTINCT group_name").
		Order("group_name asc").
		Pluck("group_name", &groups)

	if len(groups) == 0 {
		groups = []string{"默认"}
	}

	c.JSON(http.StatusOK, NewSuccessResponse(groups))
}

func (s *WebServer) healthCheck(c *gin.Context) {
	db := database.GetDB()
	sqlDB, err := db.DB()
	dbOk := err == nil && sqlDB.Ping() == nil

	c.JSON(http.StatusOK, NewSuccessResponse(map[string]interface{}{
		"status":    "ok",
		"database":  dbOk,
		"monitors":  len(monitor.GetAllMonitors()),
	}))
}

// 推送通知设置 API

func (s *WebServer) getNotificationSettings(c *gin.Context) {
	enabledVal, _ := database.GetSetting("notifications_enabled")
	enabled := enabledVal == "true"

	service, _ := database.GetSetting("notification_service")
	configRaw, _ := database.GetSetting("notification_config")

	var config map[string]interface{}
	if configRaw != "" {
		json.Unmarshal([]byte(configRaw), &config)
	}
	if config == nil {
		config = make(map[string]interface{})
	}

	c.JSON(http.StatusOK, NewSuccessResponse(map[string]interface{}{
		"enabled": enabled,
		"service": service,
		"config":  config,
	}))
}

func (s *WebServer) updateNotificationSettings(c *gin.Context) {
	var req struct {
		Enabled bool                   `json:"enabled"`
		Service string                 `json:"service"`
		Config  map[string]interface{} `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(400, "invalid request body: "+err.Error()))
		return
	}

	// 保存到数据库
	database.SetSetting("notifications_enabled", fmt.Sprintf("%t", req.Enabled))
	database.SetSetting("notification_service", req.Service)

	if req.Config != nil {
		configJSON, _ := json.Marshal(req.Config)
		database.SetSetting("notification_config", string(configJSON))
	}

	// 更新运行时状态
	notify.SetEnabled(req.Enabled)

	// 如果启用了新服务且 notifier 未初始化，尝试初始化
	if req.Enabled && notify.GetNotifier() == nil && req.Service != "" {
		notify.Reset()
		if err := notify.InitGlobalNotifier(req.Service, req.Config); err != nil {
			c.JSON(http.StatusOK, NewSuccessResponse(map[string]interface{}{
				"enabled": true,
				"warning": "服务初始化失败: " + err.Error(),
			}))
			return
		}
	}

	log.Printf("[通知] 推送设置已更新: enabled=%v service=%s", req.Enabled, req.Service)
	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}
