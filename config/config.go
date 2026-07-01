package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"
)

// FieldConfig 字段配置（用于校验和种子数据）
type FieldConfig struct {
	Name      string `json:"name"`
	Selector  string `json:"selector"`
	Attr      string `json:"attr,omitempty"`
	Type      string `json:"type"`
	Transform string `json:"transform,omitempty"`
}

// SiteSelectors 选择器配置（用于校验和种子数据）
type SiteSelectors struct {
	Container string        `json:"container"`
	Item      string        `json:"item,omitempty"`
	Fields    []FieldConfig `json:"fields"`
}

// SeedSiteConfig 种子配置中的单个站点
type SeedSiteConfig struct {
	Name          string       `json:"name"`
	URL           string       `json:"url"`
	Group         string       `json:"group,omitempty"`
	Storage       string       `json:"storage,omitempty"`
	Selectors     SiteSelectors `json:"selectors"`
	CheckInterval int          `json:"check_interval"`
}

// SeedConfig 完整的种子配置文件结构
type SeedConfig struct {
	Notification *struct {
		Service string                 `json:"service"`
		Config  map[string]interface{} `json:"config"`
	} `json:"notification,omitempty"`
	Sites []SeedSiteConfig `json:"sites"`
}

// Validate 校验选择器配置
func (ss *SiteSelectors) Validate() error {
	if ss.Container == "" {
		return errors.New("容器选择器不能为空")
	}
	for i, field := range ss.Fields {
		if field.Name == "" {
			return fmt.Errorf("字段 #%d: 名称不能为空", i+1)
		}
		if field.Selector == "" {
			return fmt.Errorf("字段 #%d: 选择器不能为空", i+1)
		}
	}
	return nil
}

// ValidateURL 校验 URL 格式
func ValidateURL(rawURL string) error {
	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return fmt.Errorf("URL格式错误 (%s)", rawURL)
	}
	return nil
}

// GetCheckInterval 计算检查间隔（秒 -> Duration）
func GetCheckInterval(seconds int) time.Duration {
	switch {
	case seconds <= 0:
		return 1 * time.Hour
	default:
		return time.Duration(seconds) * time.Second
	}
}

// LoadSeedConfig 从 JSON 文件加载种子配置（含通知和站点列表）
func LoadSeedConfig(path string) (*SeedConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg SeedConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}