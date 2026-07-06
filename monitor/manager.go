package monitor

import (
	"log"
	"sync"
	"time"

	"github.com/cn-maul/AlterBot/database"
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
	Group         string        `json:"group"`
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

	monitors[m.site.Name] = m
	log.Printf("[%s] 监控器已注册", m.site.Name)
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

// StartAllFromDB 从数据库加载所有站点并注册/启动
func StartAllFromDB() {
	var sites []database.Site
	if err := database.GetDB().Preload("Fields").Find(&sites).Error; err != nil {
		log.Printf("[Monitor] 从数据库加载站点失败: %v", err)
		return
	}

	activeCount := 0
	for i := range sites {
		if sites[i].IsActive {
			go Start(&sites[i])
			activeCount++
		} else {
			// 注册到注册表但不启动，供管理页面展示（可手动启动）
			m := NewMonitor(&sites[i])
			m.updateStatus(func(s *MonitorStatus) {
				s.IsRunning = false
			})
		}
	}
	log.Printf("[Monitor] 已启动 %d 个监控器（共 %d 个站点）", activeCount, len(sites))
}

// StopAll 停止所有正在运行的监控器
func StopAll() {
	monitorsLock.RLock()
	for name, m := range monitors {
		m.Stop()
		log.Printf("[%s] 监控器已停止", name)
	}
	monitorsLock.RUnlock()
	log.Println("[Monitor] 所有监控器已停止")
}

// StopMonitor 按名称停止单个监控器，返回是否成功
func StopMonitor(name string) bool {
	monitorsLock.RLock()
	m, exists := monitors[name]
	monitorsLock.RUnlock()
	if !exists {
		return false
	}
	m.Stop()
	log.Printf("[%s] 监控器已停止", name)
	return true
}

// Exists 检查监控器是否存在
func Exists(name string) bool {
	monitorsLock.RLock()
	defer monitorsLock.RUnlock()
	_, ok := monitors[name]
	return ok
}

// MarkRead 将指定监控器的未读计数归零（用户已查看详情后调用）
func MarkRead(name string) {
	monitorsLock.RLock()
	m, exists := monitors[name]
	monitorsLock.RUnlock()
	if !exists {
		return
	}
	m.updateStatus(func(s *MonitorStatus) {
		s.UpdatesCount = 0
	})
}
