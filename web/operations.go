package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cn-maul/AlterBot/database"
	"github.com/cn-maul/AlterBot/monitor"
	"github.com/cn-maul/AlterBot/notify"
	"github.com/gin-gonic/gin"
)

// addMonitorRequest 创建监控器的请求体
type addMonitorRequest struct {
	Name           string         `json:"name" binding:"required"`
	URL            string         `json:"url" binding:"required"`
	Container      string         `json:"container" binding:"required"`
	Item           string         `json:"item"`
	Group          string         `json:"group"`
	CheckInterval  int            `json:"check_interval"`
	IsActive       bool           `json:"is_active"`
	NotifyFilter   string         `json:"notify_filter"`
	NotifyKeywords string         `json:"notify_keywords"`
	NotifyAccountIDs string       `json:"notify_account_ids"`
	Fields         []fieldRequest `json:"fields"`
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
		NotifyFilter:   req.NotifyFilter,
		NotifyKeywords: req.NotifyKeywords,
		NotifyAccountIDs: req.NotifyAccountIDs,
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
	// 从注册表移除
	monitor.UnregisterMonitor(name)

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
	oldName := c.Param("name")

	var req addMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(400, "invalid request body: "+err.Error()))
		return
	}

	// 从数据库查找
	var site database.Site
	if err := database.GetDB().Where("name = ?", oldName).First(&site).Error; err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse(404, "monitor not found"))
		return
	}

	// 停止正在运行的旧实例（不注销，稍后复用或替换）
	oldMonitor := monitor.GetMonitor(oldName)
	if oldMonitor != nil && oldMonitor.GetStatus().IsRunning {
		oldMonitor.Stop()
	}

	// 如果改名，把旧名从注册表移除（新名由 startMonitorGoroutine 注册）
	nameChanged := req.Name != "" && req.Name != oldName
	if nameChanged {
		monitor.UnregisterMonitor(oldName)
	}

	// 更新数据库
	group := req.Group
	if group == "" {
		group = "默认"
	}
	site.Name = req.Name
	site.URL = req.URL
	site.Container = req.Container
	site.Item = req.Item
	site.GroupName = group
	site.CheckInterval = req.CheckInterval
	site.IsActive = req.IsActive
	site.NotifyFilter = req.NotifyFilter
	site.NotifyKeywords = req.NotifyKeywords
	site.NotifyAccountIDs = req.NotifyAccountIDs
	database.GetDB().Save(&site)

	// 重设字段
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

	// 如果要求启动，重新读取完整配置并启动
	if req.IsActive {
		var updatedSite database.Site
		database.GetDB().Preload("Fields").First(&updatedSite, site.ID)
		startMonitorGoroutine(&updatedSite)
	}

	newName := req.Name
	if newName == "" {
		newName = oldName
	}
	log.Printf("[Web] 更新监控器: %s -> %s", oldName, newName)
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

func (s *WebServer) markAllNotified(c *gin.Context) {
	name := c.Param("name")

	var site database.Site
	if err := database.GetDB().Where("name = ?", name).First(&site).Error; err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse(404, "monitor not found"))
		return
	}

	now := time.Now()
	result := database.GetDB().Model(&database.UpdateRecord{}).
		Where("site_id = ? AND notified = ?", site.ID, false).
		Updates(map[string]interface{}{
			"notified":     true,
			"notified_at": now,
		})

	log.Printf("[Web] 标记 %s 的 %d 条记录为已推送", name, result.RowsAffected)
	c.JSON(http.StatusOK, NewSuccessResponse(map[string]interface{}{
		"updated": result.RowsAffected,
	}))
}

func (s *WebServer) markRead(c *gin.Context) {
	name := c.Param("name")
	monitor.MarkRead(name)
	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}

func (s *WebServer) updateNotifyAccounts(c *gin.Context) {
	name := c.Param("name")
	var req struct {
		AccountIDs string `json:"notify_account_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(400, "参数错误: "+err.Error()))
		return
	}

	var site database.Site
	if err := database.GetDB().Where("name = ?", name).First(&site).Error; err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse(404, "monitor not found"))
		return
	}

	site.NotifyAccountIDs = req.AccountIDs
	database.GetDB().Save(&site)

	// 同步更新运行中的监控器实例（避免下次检查周期仍用旧配置）
	if m := monitor.GetMonitor(name); m != nil {
		m.UpdateSiteNotifyAccounts(site.NotifyAccountIDs)
	}

	log.Printf("[Web] 更新 %s 的推送账户: %s", name, req.AccountIDs)
	c.JSON(http.StatusOK, NewSuccessResponse(nil))
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

// ===== 推送账户 CRUD =====

type accountRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Service     string                 `json:"service" binding:"required"`
	Config      map[string]interface{} `json:"config" binding:"required"`
}

func (s *WebServer) listAccounts(c *gin.Context) {
	var accounts []database.NotificationAccount
	database.GetDB().Order("created_at desc").Find(&accounts)
	c.JSON(http.StatusOK, NewSuccessResponse(accounts))
}

func (s *WebServer) createAccount(c *gin.Context) {
	var req accountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(400, "参数错误: "+err.Error()))
		return
	}

	configJSON, _ := json.Marshal(req.Config)
	account := &database.NotificationAccount{
		Name:        req.Name,
		Service:     req.Service,
		ConfigJSON:  string(configJSON),
	}
	if err := database.GetDB().Create(account).Error; err != nil {
		c.JSON(http.StatusConflict, NewErrorResponse(409, "创建账户失败: "+err.Error()))
		return
	}

	log.Printf("[通知] 创建推送账户: %s (%s)", account.Name, account.Service)
	c.JSON(http.StatusCreated, NewSuccessResponse(account))
}

func (s *WebServer) updateAccount(c *gin.Context) {
	var req accountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(400, "参数错误: "+err.Error()))
		return
	}

	id := c.Param("id")
	var account database.NotificationAccount
	if err := database.GetDB().First(&account, id).Error; err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse(404, "账户不存在"))
		return
	}

	configJSON, _ := json.Marshal(req.Config)
	account.Name = req.Name
	account.Service = req.Service
	account.ConfigJSON = string(configJSON)
	database.GetDB().Save(&account)

	log.Printf("[通知] 更新推送账户: %s", account.Name)
	c.JSON(http.StatusOK, NewSuccessResponse(account))
}

func (s *WebServer) deleteAccount(c *gin.Context) {
	id := c.Param("id")
	result := database.GetDB().Delete(&database.NotificationAccount{}, id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, NewErrorResponse(404, "账户不存在"))
		return
	}
	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}

// 推送全局开关

func (s *WebServer) getNotificationSettings(c *gin.Context) {
	enabledVal, _ := database.GetSetting("notifications_enabled")
	enabled := enabledVal == "true"

	c.JSON(http.StatusOK, NewSuccessResponse(map[string]interface{}{
		"enabled": enabled,
	}))
}

func (s *WebServer) updateNotificationSettings(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(400, "参数错误: "+err.Error()))
		return
	}

	database.SetSetting("notifications_enabled", fmt.Sprintf("%t", req.Enabled))
	notify.SetEnabled(req.Enabled)

	log.Printf("[通知] 推送开关已更新: enabled=%v", req.Enabled)
	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}

// ===== 智能扫描 =====

func (s *WebServer) previewScan(c *gin.Context) {
	var req struct {
		URL      string `json:"url" binding:"required"`
		Keywords string `json:"keywords" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(400, "参数错误: "+err.Error()))
		return
	}

	// 解析关键词
	var keywords []string
	for _, kw := range splitKeywords(req.Keywords) {
		if kw != "" {
			keywords = append(keywords, kw)
		}
	}
	if len(keywords) == 0 {
		c.JSON(http.StatusBadRequest, NewErrorResponse(400, "至少需要一个关键词"))
		return
	}

	result, err := monitor.SmartScan(&monitor.ScanSettings{
		URL:      req.URL,
		Keywords: keywords,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(500, "扫描失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(result))
}

func (s *WebServer) smartCreate(c *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		URL          string `json:"url" binding:"required"`
		ContainerCSS string `json:"container_css" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(400, "参数错误: "+err.Error()))
		return
	}

	_, err := monitor.MonitorFromScan(req.Name, req.URL, req.ContainerCSS)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(500, "创建失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusCreated, NewSuccessResponse(map[string]interface{}{
		"name": req.Name,
	}))
}

// splitKeywords 分割关键词（支持中英文逗号、空格）
func splitKeywords(s string) []string {
	var result []string
	current := ""
	for _, r := range s {
		if r == ',' || r == '，' || r == '　' || r == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
