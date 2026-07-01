package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cn-maul/AlterBot/config"
	"github.com/cn-maul/AlterBot/database"
	"github.com/cn-maul/AlterBot/monitor"
	"github.com/cn-maul/AlterBot/notify"
	"github.com/cn-maul/AlterBot/web"
)

//go:embed frontend/dist
var frontendDist embed.FS

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("=== AlterBot 网页变更监控系统 ===")

	// 1. 初始化数据库
	dbPath := "alterbot.db"
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 2. 加载用户配置（如 config.json 存在），导入种子站点 + 初始化推送
	tryLoadConfig()

	// 3. 从数据库加载并启动所有活跃的监控器
	monitor.StartAllFromDB()

	// 4. 载入前端（如果已构建）
	var frontendFS fs.FS
	if _, err := fs.ReadDir(frontendDist, "frontend/dist"); err == nil {
		sub, err := fs.Sub(frontendDist, "frontend/dist")
		if err == nil {
			frontendFS = sub
			log.Println("[Web] 前端已嵌入，管理界面可用")
		}
	} else {
		log.Println("[Web] 前端未构建，仅 API 模式运行（cd frontend && npm run dev）")
	}

	// 5. 启动 Web 服务
	ws := web.NewWebServer(frontendFS)
	go func() {
		addr := ":8080"
		log.Printf("[Web] 服务启动: http://localhost%s", addr)
		if err := ws.Run(addr); err != nil {
			log.Fatalf("Web 服务启动失败: %v", err)
		}
	}()

	// 5. 等待中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("收到信号 %v，正在停止所有监控器...", sig)

	monitor.StopAll()
	log.Println("AlterBot 已安全退出")
}

// tryLoadConfig 尝试加载 config.json，初始化推送+种子站点
func tryLoadConfig() {
	cfg, err := config.LoadSeedConfig("config.json")
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("[配置] 未找到 config.json，使用空配置启动")
			return
		}
		log.Printf("[配置] 加载 config.json 失败: %v", err)
		return
	}

	// 初始化推送服务（从种子配置）
	if cfg.Notification != nil {
		if err := notify.InitGlobalNotifier(
			cfg.Notification.Service,
			cfg.Notification.Config,
		); err != nil {
			log.Printf("[配置] 推送服务初始化失败: %v", err)
		} else {
			log.Printf("[配置] 推送服务已注册: %s（默认关闭，需在管理页面启用）", cfg.Notification.Service)
		}
		// 保存通知配置到数据库，供管理页面使用
		configJSON, _ := json.Marshal(cfg.Notification.Config)
		database.SetSetting("notification_service", cfg.Notification.Service)
		database.SetSetting("notification_config", string(configJSON))
		// 检查用户是否在管理页面启用了推送
		if enabledVal, ok := database.GetSetting("notifications_enabled"); ok && enabledVal == "true" {
			notify.SetEnabled(true)
		}
	}

	// 导入种子站点（仅当新站点时创建，已存在则跳过）
	for _, site := range cfg.Sites {
		group := site.Group
		if group == "" {
			group = "默认"
		}

		// 检查是否已存在
		var count int64
		database.GetDB().Model(&database.Site{}).Where("name = ?", site.Name).Count(&count)
		if count > 0 {
			log.Printf("[配置] 站点「%s」已存在，跳过导入", site.Name)
			continue
		}

		// 构建 Site 记录
		dbSite := &database.Site{
			Name:          site.Name,
			URL:           site.URL,
			Container:     site.Selectors.Container,
			Item:          site.Selectors.Item,
			GroupName:     group,
			CheckInterval: site.CheckInterval,
			IsActive:      true,
		}
		for _, f := range site.Selectors.Fields {
			ft := f.Type
			if ft == "" {
				ft = "text"
			}
			dbSite.Fields = append(dbSite.Fields, database.SiteField{
				Name:      f.Name,
				Selector:  f.Selector,
				Type:      ft,
				Attr:      f.Attr,
				Transform: f.Transform,
			})
		}

		if err := database.GetDB().Create(dbSite).Error; err != nil {
			log.Printf("[配置] 导入站点「%s」失败: %v", site.Name, err)
		} else {
			log.Printf("[配置] 导入站点「%s」到分组「%s」", site.Name, group)
		}
	}
}