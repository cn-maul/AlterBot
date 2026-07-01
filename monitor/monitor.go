package monitor

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/cn-maul/AlterBot/database"
	"github.com/cn-maul/AlterBot/fetcher"
	"github.com/cn-maul/AlterBot/notify"
)

type Monitor struct {
	site       *database.Site
	extractor  *Extractor
	fetcher    *fetcher.Fetcher
	stopCh     chan struct{}
	status     MonitorStatus
	statusLock sync.RWMutex
}

func NewMonitor(site *database.Site, fetcherOpts ...fetcher.Option) *Monitor {
	f := fetcher.New(fetcherOpts...)

	// 从 database.Site 构建选择器信息
	selectors := SiteSelectors{
		Container: site.Container,
		Item:      site.Item,
		Fields:    make([]FieldConfig, len(site.Fields)),
	}
	for i, f := range site.Fields {
		selectors.Fields[i] = FieldConfig{
			Name:      f.Name,
			Selector:  f.Selector,
			Type:      f.Type,
			Attr:      f.Attr,
			Transform: f.Transform,
		}
	}

	m := &Monitor{
		site:      site,
		extractor: NewExtractor(selectors),
		fetcher:   f,
		stopCh:    make(chan struct{}),
		status: MonitorStatus{
			Name:          site.Name,
			URL:           site.URL,
			Group:         site.GroupName,
			IsRunning:     true,
			CheckInterval: site.GetCheckInterval(),
			NextCheck:     time.Now().Add(site.GetCheckInterval()),
		},
	}

	RegisterMonitor(m)
	return m
}

func Start(site *database.Site) {
	m := NewMonitor(site)
	ticker := time.NewTicker(site.GetCheckInterval())
	defer ticker.Stop()

	m.updateStatus(func(s *MonitorStatus) {
		s.IsRunning = true
		s.NextCheck = time.Now().Add(site.GetCheckInterval())
	})

	log.Printf("[%s] 监控启动，检查间隔: %v", site.Name, site.GetCheckInterval())
	performCheck(m, true) // 首次检查

	for {
		select {
		case <-m.stopCh:
			m.updateStatus(func(s *MonitorStatus) {
				s.IsRunning = false
			})
			log.Printf("[%s] 监控停止", site.Name)
			return
		case <-ticker.C:
			performCheck(m, false)
		}
	}
}

func performCheck(m *Monitor, isFirst bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[%s] 检查异常: %v", m.site.Name, r)
			m.updateStatus(func(s *MonitorStatus) {
				s.LastError = fmt.Sprintf("panic: %v", r)
			})
		}
	}()

	startTime := time.Now()
	updates, err := m.CheckForUpdates()
	duration := time.Since(startTime)

	updateMonitorStatus(m, updates, err, duration)
	logCheckResult(m, updates, err, duration, isFirst)

	if len(updates) > 0 {
		m.sendCombinedNotification(updates)
	}
}

func updateMonitorStatus(m *Monitor, updates []ExtractResult, err error, duration time.Duration) {
	m.updateStatus(func(s *MonitorStatus) {
		s.LastCheck = time.Now()
		s.LastDuration = duration
		s.NextCheck = time.Now().Add(m.site.GetCheckInterval())

		if err != nil {
			s.LastError = err.Error()
		} else {
			s.LastError = ""
			if len(updates) > 0 {
				s.LastUpdate = time.Now()
				s.UpdatesCount += len(updates)
			}
		}
	})

	// 同步更新数据库中的 last_check_at
	database.GetDB().Model(m.site).Update("LastCheckAt", time.Now())
}

func logCheckResult(m *Monitor, updates []ExtractResult, err error, duration time.Duration, isFirst bool) {
	prefix := "检查"
	if isFirst {
		prefix = "首次检查"
	}

	name := m.site.Name
	if err != nil {
		log.Printf("[%s] %s失败 (耗时: %v): %v", name, prefix, duration, err)
		return
	}

	if len(updates) > 0 {
		log.Printf("[%s] %s发现 %d 条更新 (耗时: %v)", name, prefix, len(updates), duration)
		for _, item := range updates {
			log.Printf(" - %s", item["title"])
		}
	} else {
		log.Printf("[%s] %s未发现新内容 (耗时: %v)", name, prefix, duration)
	}
}

func (m *Monitor) CheckForUpdates() ([]ExtractResult, error) {
	html, err := m.fetcher.Fetch(m.site.URL)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}

	current, err := m.extractor.Extract(html)
	if err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}

	last, err := m.loadLastResults()
	if err != nil {
		return nil, fmt.Errorf("load history failed: %w", err)
	}

	newItems := compareResults(last, current)

	if err := m.saveResults(current); err != nil {
		return nil, fmt.Errorf("save failed: %w", err)
	}

	// 将新变更保存到数据库更新记录
	for _, item := range newItems {
		record := &database.UpdateRecord{
			SiteID:  m.site.ID,
			Title:   toString(item["title"]),
			URL:     toString(item["url"]),
			Content: func() string { data, _ := json.Marshal(item); return string(data) }(),
		}
		database.GetDB().Create(record)
	}

	return newItems, nil
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func (m *Monitor) loadLastResults() ([]ExtractResult, error) {
	// 从数据库读取最近一次的提取结果
	var records []database.UpdateRecord
	if err := database.GetDB().Where("site_id = ?", m.site.ID).
		Order("created_at desc").Limit(50).Find(&records).Error; err != nil {
		return nil, nil
	}

	if len(records) == 0 {
		return nil, nil
	}

	var results []ExtractResult
	for _, r := range records {
		if r.Content != "" {
			var item ExtractResult
			if err := json.Unmarshal([]byte(r.Content), &item); err == nil {
				results = append(results, item)
			}
		}
	}
	return results, nil
}

func (m *Monitor) saveResults(results []ExtractResult) error {
	if len(results) == 0 {
		return nil
	}

	// 将提取结果存入数据库
	for _, item := range results {
		data, _ := json.Marshal(item)
		record := &database.UpdateRecord{
			SiteID:  m.site.ID,
			Title:   toString(item["title"]),
			URL:     toString(item["url"]),
			Content: string(data),
		}
		// 检查是否已存在（避免重复）
		var count int64
		database.GetDB().Model(&database.UpdateRecord{}).
			Where("site_id = ? AND title = ? AND url = ?", m.site.ID, record.Title, record.URL).
			Count(&count)
		if count == 0 {
			database.GetDB().Create(record)
		}
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

func (m *Monitor) sendCombinedNotification(items []ExtractResult) {
	if !notify.IsEnabled() {
		log.Printf("[%s] 推送已关闭，跳过 %d 条通知", m.site.Name, len(items))
		return
	}
	if notifier := notify.GetNotifier(); notifier != nil {
		title := fmt.Sprintf("%s 有 %d 条更新", m.site.Name, len(items))

		var content strings.Builder
		content.WriteString("最新更新内容：\n")
		for i, item := range items {
			fmt.Fprintf(&content, "%d. %s\n   %s\n", i+1, item["title"], item["url"])
		}

		if err := notifier.Send(title, content.String()); err != nil {
			log.Printf("[%s] 推送失败: %v", m.site.Name, err)
			return
		}

		// 推送成功后标记数据库记录为已通知
		now := time.Now()
		for _, item := range items {
			title := toString(item["title"])
			urlStr := toString(item["url"])
			database.GetDB().Model(&database.UpdateRecord{}).
				Where("site_id = ? AND title = ? AND url = ? AND notified = ?", m.site.ID, title, urlStr, false).
				Updates(map[string]interface{}{
					"notified":     true,
					"notified_at": now,
				})
		}
		log.Printf("[%s] 推送成功，已标记 %d 条记录", m.site.Name, len(items))
	}
}

func extractKey(item ExtractResult) string {
	title, _ := item["title"].(string)
	urlStr, _ := item["url"].(string)
	switch {
	case title != "" && urlStr != "":
		return title + "|" + urlStr
	case title != "":
		return title
	case urlStr != "":
		return urlStr
	default:
		data, err := json.Marshal(item)
		if err == nil {
			return fmt.Sprintf("%x", data)
		}
		return ""
	}
}
