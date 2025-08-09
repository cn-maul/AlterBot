package monitor

import (
	"encoding/json/v2"
	"fmt"
	"github.com/cn-maul/AlterBot/config"
	"github.com/cn-maul/AlterBot/fetcher"
	"github.com/cn-maul/AlterBot/notify"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type Monitor struct {
	config     *config.SiteConfig
	storage    Storage
	extractor  *Extractor
	fetcher    *fetcher.Fetcher
	stopCh     chan struct{}
	status     MonitorStatus
	statusLock sync.RWMutex
}

func NewMonitor(siteConfig *config.SiteConfig, fetcherOpts ...fetcher.Option) *Monitor {
	f := fetcher.New(append([]fetcher.Option{}, fetcherOpts...)...)

	m := &Monitor{
		config:    siteConfig,
		storage:   &FileStorage{FilePath: siteConfig.Storage},
		extractor: NewExtractor(siteConfig.Selectors),
		fetcher:   f,
		stopCh:    make(chan struct{}),
		status: MonitorStatus{
			Name:          siteConfig.Name,
			URL:           siteConfig.URL,
			IsRunning:     true,
			CheckInterval: siteConfig.GetCheckInterval(),
			NextCheck:     time.Now().Add(siteConfig.GetCheckInterval()),
		},
	}

	RegisterMonitor(m)
	return m
}

func Start(site *config.SiteConfig, stopCh <-chan struct{}) {
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
		case <-stopCh:
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
		s.NextCheck = time.Now().Add(m.config.GetCheckInterval())

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
}

func logCheckResult(m *Monitor, updates []ExtractResult, err error, duration time.Duration, isFirst bool) {
	prefix := "检查"
	if isFirst {
		prefix = "首次检查"
	}

	if err != nil {
		log.Printf("[%s] %s失败 (耗时: %v): %v", m.config.Name, prefix, duration, err)
		return
	}

	if len(updates) > 0 {
		log.Printf("[%s] %s发现 %d 条更新 (耗时: %v)", m.config.Name, prefix, len(updates), duration)
		for _, item := range updates {
			log.Printf(" - %s", item["title"])
		}
	} else {
		log.Printf("[%s] %s未发现新内容 (耗时: %v)", m.config.Name, prefix, duration)
	}
}

func (m *Monitor) CheckForUpdates() ([]ExtractResult, error) {
	html, err := m.fetcher.Fetch(m.config.URL)
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

	return newItems, nil
}

func (m *Monitor) loadLastResults() ([]ExtractResult, error) {
	data, err := m.storage.Load()
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
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
	opts := json.DefaultOptionsV2()
	data, err := json.Marshal(results, opts)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}

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

func (m *Monitor) sendCombinedNotification(items []ExtractResult) {
	if notifier := notify.GetNotifier(); notifier != nil {
		title := fmt.Sprintf("%s 有 %d 条更新", m.config.Name, len(items))

		var content strings.Builder
		content.WriteString("最新更新内容：\n")
		for i, item := range items {
			fmt.Fprintf(&content, "%d. %s\n   %s\n", i+1, item["title"], item["url"])
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
