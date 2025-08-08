package notify

import "sync"

var (
	providers     = make(map[string]func(map[string]interface{}) (Notifier, error))
	providersLock sync.RWMutex
)

// Register 注册推送服务
func Register(name string, creator func(map[string]interface{}) (Notifier, error)) {
	providersLock.Lock()
	defer providersLock.Unlock()
	providers[name] = creator
}
