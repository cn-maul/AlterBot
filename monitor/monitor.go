package monitor

import (
	"AlterBot/config"
	"AlterBot/fetcher"
	"AlterBot/notify"
	"encoding/json/v2"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Monitor struct {
	config    *config.SiteConfig
	storage   Storage
	extractor *Extractor
	fetcher   *fetcher.Fetcher
	stopCh    chan struct{}
}

func NewMonitor(siteConfig *config.SiteConfig, fetcherOpts ...fetcher.Option) *Monitor {
	// 创建 fetcher 实例（允许传入自定义配置）
	f := fetcher.New(
		append([]fetcher.Option{}, fetcherOpts...)...,
	)

	return &Monitor{
		config:    siteConfig,
		storage:   &FileStorage{FilePath: siteConfig.Storage},
		extractor: NewExtractor(siteConfig.Selectors),
		fetcher:   f,
		stopCh:    make(chan struct{}),
	}
}

// 单个监控器运行逻辑
func Start(site *config.SiteConfig, stopCh <-chan struct{}) {
	m := NewMonitor(site)
	ticker := time.NewTicker(site.GetCheckInterval())
	defer ticker.Stop()

	log.Printf("[%s] 监控启动，间隔: %v", site.Name, site.GetCheckInterval())

	// 立即执行首次检查
	startTime := time.Now()
	log.Printf("[%s] 开始首次检查...", site.Name)
	if updates, err := m.CheckForUpdates(); err != nil {
		log.Printf("[%s] 首次检查失败: %v", site.Name, err)
	} else if len(updates) > 0 {
		log.Printf("[%s] 首次检查发现 %d 条更新", site.Name, len(updates))
		// 这里可以添加简单的通知逻辑
		m.sendCombinedNotification(updates)
	} else {
		log.Printf("[%s] 首次检查未发现新内容", site.Name)
	}
	log.Printf("[%s] 首次检查完成 (耗时: %v)", site.Name, time.Since(startTime))

	// 定时检查循环
	for {
		select {
		case <-stopCh:
			log.Printf("[%s] 监控停止", site.Name)
			return
		case <-ticker.C:
			startTime := time.Now()
			log.Printf("[%s] 开始检查更新...", site.Name)

			updates, err := m.CheckForUpdates()
			elapsed := time.Since(startTime)

			if err != nil {
				log.Printf("[%s] 检查失败 (耗时: %v): %v", site.Name, elapsed, err)
				continue
			}

			if len(updates) > 0 {
				log.Printf("[%s] 发现 %d 条更新 (耗时: %v)", site.Name, len(updates), elapsed)
				// 这里可以添加简单的通知逻辑
				m.sendCombinedNotification(updates)
				for _, item := range updates {
					log.Printf(" - %s", item["title"])
				}
			} else {
				log.Printf("[%s] 没有发现新内容 (耗时: %v)", site.Name, elapsed)
			}
		}
	}
}

func (m *Monitor) CheckForUpdates() ([]ExtractResult, error) {
	// 1. 获取网页内容
	html, err := m.fetcher.Fetch(m.config.URL)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}

	// 2. 提取当前内容
	current, err := m.extractor.Extract(html)
	if err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}

	// 3. 加载上次结果
	last, err := m.loadLastResults()
	if err != nil {
		return nil, fmt.Errorf("load history failed: %w", err)
	}

	// 4. 比较差异
	newItems := compareResults(last, current)

	// 6. 保存当前结果
	if err := m.saveResults(current); err != nil {
		return nil, fmt.Errorf("save failed: %w", err)
	}

	return newItems, nil
}

func (m *Monitor) loadLastResults() ([]ExtractResult, error) {
	data, err := m.storage.Load()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 首次运行无历史数据
		}
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	var results []ExtractResult
	opts := json.DefaultOptionsV2()
	if err := json.Unmarshal(data, &results, opts); err != nil {
		return nil, err
	}

	return results, nil
}

func (m *Monitor) saveResults(results []ExtractResult) error {
	// 使用 json/v2 的 MarshalOptions 进行更灵活的配置
	opts := json.DefaultOptionsV2()

	// 使用 Marshal 方法序列化数据
	data, err := json.Marshal(results, opts)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}

	// 保存到存储
	if err := m.storage.Save(data); err != nil {
		return fmt.Errorf("storage save failed: %w", err)
	}

	return nil
}

func compareResults(last, current []ExtractResult) []ExtractResult {
	lastKeys := make(map[string]struct{})
	for _, item := range last {
		if key := extractKey(item); key != "" {
			lastKeys[key] = struct{}{}
		}
	}

	var newItems []ExtractResult
	for _, item := range current {
		if key := extractKey(item); key != "" {
			if _, exists := lastKeys[key]; !exists {
				newItems = append(newItems, item)
			}
		}
	}

	return newItems
}

// sendCombinedNotification 合并多条更新为单次推送
func (m *Monitor) sendCombinedNotification(items []ExtractResult) {
	if notifier := notify.GetNotifier(); notifier != nil {
		title := fmt.Sprintf("%s 有 %d 条更新", m.config.Name, len(items))

		var content strings.Builder
		content.WriteString("最新更新内容：\n")
		for i, item := range items {
			fmt.Fprintf(&content, "%d. %s\n   %s\n",
				i+1,
				item["title"],
				item["url"])
		}

		if err := notifier.Send(title, content.String()); err != nil {
			log.Printf("[%s] 推送失败: %v", m.config.Name, err)
		}
	}
}

func extractKey(item ExtractResult) string {
	if key, ok := item["title"].(string); ok {
		return key
	}
	if key, ok := item["name"].(string); ok {
		return key
	}
	return ""
}

/*
func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 创建WaitGroup和关闭通道
	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	// 3. 为每个站点启动监控goroutine
	for _, site := range cfg.Sites {
		wg.Go(func() { // Go 1.25新语法
			runMonitor(&site, stopCh)
		})
	}

	// 4. 等待中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// 5. 关闭所有监控器
	close(stopCh)
	wg.Wait()
	log.Println("所有监控器已停止")
}

// 单个监控器运行逻辑
func runMonitor(site *config.SiteConfig, stopCh <-chan struct{}) {
	m := monitor.NewMonitor(site)
	ticker := time.NewTicker(site.GetCheckInterval())
	defer ticker.Stop()

	log.Printf("[%s] 监控启动，间隔: %v", site.Name, site.GetCheckInterval())

	for {
		select {
		case <-stopCh:
			log.Printf("[%s] 监控停止", site.Name)
			return
		case <-ticker.C:
			updates, err := m.CheckForUpdates()
			if err != nil {
				log.Printf("[%s] 监控错误: %v", site.Name, err)
				continue
			}

			if len(updates) > 0 {
				log.Printf("[%s] 发现 %d 条更新", site.Name, len(updates))
				// 这里可以添加简单的通知逻辑
				for _, item := range updates {
					log.Printf(" - %s", item["title"])
				}
			}
		}
	}
}
*/
