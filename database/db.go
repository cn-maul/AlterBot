package database

import (
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化 SQLite 数据库并自动迁移
func Init(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	// 自动迁移 Schema
	if err := DB.AutoMigrate(&Site{}, &SiteField{}, &UpdateRecord{}, &NotificationAccount{}, &SystemSetting{}); err != nil {
		return err
	}

	log.Printf("[DB] 数据库就绪: %s", dbPath)
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// Now 返回当前时间（统一时间源）
func Now() time.Time {
	return time.Now()
}

// GetSetting 获取系统设置
func GetSetting(key string) (string, bool) {
	var s SystemSetting
	DB.Where("key = ?", key).Limit(1).Find(&s)
	if s.Key == "" {
		return "", false
	}
	return s.Value, true
}

// SetSetting 设置系统设置
func SetSetting(key, value string) error {
	return DB.Where("key = ?", key).Assign(SystemSetting{Value: value}).FirstOrCreate(&SystemSetting{Key: key}).Error
}