package notify

import (
	"fmt"
	"sync"
)

var (
	globalNotifier Notifier
	initOnce       sync.Once
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
