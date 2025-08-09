package monitor

import (
	"log"
	"sync"
	"time"
)

// 全局监控器注册表
var (
	monitors     = make(map[string]*Monitor)
	monitorsLock sync.RWMutex
)

// MonitorStatus 表示监控器的状态信息
type MonitorStatus struct {
	Name          string        `json:"name"`
	URL           string        `json:"url"`
	IsRunning     bool          `json:"is_running"`
	LastCheck     time.Time     `json:"last_check"`
	LastDuration  time.Duration `json:"last_duration"`
	LastError     string        `json:"last_error,omitempty"`
	LastUpdate    time.Time     `json:"last_update,omitempty"`
	UpdatesCount  int           `json:"updates_count"`
	NextCheck     time.Time     `json:"next_check"`
	CheckInterval time.Duration `json:"check_interval"`
}

// RegisterMonitor 注册一个新的监控器到全局
func RegisterMonitor(m *Monitor) {
	monitorsLock.Lock()
	defer monitorsLock.Unlock()

	monitors[m.config.Name] = m
	log.Printf("[%s] 监控器已注册", m.config.Name)
}

// UnregisterMonitor 从全局注销监控器
func UnregisterMonitor(name string) {
	monitorsLock.Lock()
	defer monitorsLock.Unlock()

	if _, exists := monitors[name]; exists {
		delete(monitors, name)
		log.Printf("[%s] 监控器已注销", name)
	}
}

// GetMonitor 获取指定名称的监控器
func GetMonitor(name string) *Monitor {
	monitorsLock.RLock()
	defer monitorsLock.RUnlock()

	return monitors[name]
}

// GetAllMonitors 获取所有监控器的状态
func GetAllMonitors() []MonitorStatus {
	monitorsLock.RLock()
	defer monitorsLock.RUnlock()

	var statusList []MonitorStatus
	for _, m := range monitors {
		statusList = append(statusList, m.GetStatus())
	}
	return statusList
}

// GetStatus 获取监控器的当前状态
func (m *Monitor) GetStatus() MonitorStatus {
	m.statusLock.RLock()
	defer m.statusLock.RUnlock()
	return m.status
}

// updateStatus 更新监控器状态（线程安全）
func (m *Monitor) updateStatus(updater func(*MonitorStatus)) {
	m.statusLock.Lock()
	defer m.statusLock.Unlock()
	updater(&m.status)
}

/*
// 获取单个监控器状态
status := GetMonitor("example-site").GetStatus()

// 获取所有监控器状态
allStatus := GetAllMonitors()

// 输出状态信息示例
fmt.Printf(`
Name: %s
URL: %s
Running: %v
Last Check: %v (took %v)
Next Check: %v
Updates: %d
Last Error: %s
`,
	status.Name,
	status.URL,
	status.IsRunning,
	status.LastCheck.Format(time.RFC3339),
	status.LastDuration,
	status.NextCheck.Format(time.RFC3339),
	status.UpdatesCount,
	status.LastError,
)
*/
