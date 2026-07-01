package notify

import (
	"fmt"
	"sync"
)

var (
	globalNotifier  Notifier
	initOnce        sync.Once
	notificationsOn bool = false // 默认关闭推送
	settingsLock    sync.RWMutex
)

// InitGlobalNotifier 初始化全局推送服务
func InitGlobalNotifier(serviceName string, config map[string]interface{}) error {
	var initErr error
	initOnce.Do(func() {
		if creator, ok := providers[serviceName]; ok {
			globalNotifier, initErr = creator(config)
		} else {
			initErr = fmt.Errorf("未注册的推送服务: %s", serviceName)
		}
	})
	return initErr
}

// GetNotifier 获取全局推送实例
func GetNotifier() Notifier {
	return globalNotifier
}

// SetEnabled 设置推送开关
func SetEnabled(enabled bool) {
	settingsLock.Lock()
	defer settingsLock.Unlock()
	notificationsOn = enabled
}

// IsEnabled 推送是否开启
func IsEnabled() bool {
	settingsLock.RLock()
	defer settingsLock.RUnlock()
	return notificationsOn && globalNotifier != nil
}

// Reset 重置单例（用于测试/重新配置）
func Reset() {
	initOnce = sync.Once{}
	globalNotifier = nil
	notificationsOn = false
}
