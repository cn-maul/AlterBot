package database

import "time"

// Site 监控站点配置
type Site struct {
	ID        uint   `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string `gorm:"uniqueIndex;size:255"`
	URL       string `gorm:"size:512"`
	Container string `gorm:"size:255"`
	Item      string `gorm:"size:255"`
	GroupName string `gorm:"size:100;default:默认;index"`
	// CheckInterval 检查间隔（秒），默认 3600
	CheckInterval int        `gorm:"default:3600"`
	IsActive      bool       `gorm:"default:false"`
	LastCheckAt   *time.Time `gorm:"index"`
	Fields        []SiteField
}

// SiteField 提取字段配置
type SiteField struct {
	ID        uint   `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	SiteID    uint   `gorm:"index"`
	Name      string `gorm:"size:100"`
	Selector  string `gorm:"size:255"`
	Type      string `gorm:"size:20;default:text"`
	Attr      string `gorm:"size:50"`
	Transform string `gorm:"size:255"`
}

// UpdateRecord 变更历史记录
type UpdateRecord struct {
	ID         uint       `gorm:"primarykey"`
	CreatedAt  time.Time  `gorm:"index"`
	SiteID     uint       `gorm:"index"`
	Title      string     `gorm:"size:500"`
	URL        string     `gorm:"size:512"`
	Summary    string     `gorm:"size:1000"`
	Content    string     `gorm:"type:text"`
	Notified   bool       `gorm:"default:false"`
	NotifiedAt *time.Time `gorm:"index"`
	IsRead     bool       `gorm:"default:false"`
}

// TableName 自定义表名
func (Site) TableName() string { return "sites" }
func (SiteField) TableName() string { return "site_fields" }
func (UpdateRecord) TableName() string { return "update_records" }

// SystemSetting 系统设置键值对
type SystemSetting struct {
	ID    uint   `gorm:"primarykey"`
	Key   string `gorm:"uniqueIndex;size:100"`
	Value string `gorm:"type:text"`
}

func (SystemSetting) TableName() string { return "system_settings" }

// GetCheckInterval 返回 time.Duration 形式的检查间隔
func (s *Site) GetCheckInterval() time.Duration {
	switch {
	case s.CheckInterval <= 0:
		return 1 * time.Hour
	default:
		return time.Duration(s.CheckInterval) * time.Second
	}
}