package config

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"
)

type FieldConfig struct {
	Name      string `json:"name"`
	Selector  string `json:"selector"`
	Attr      string `json:"attr,omitempty"`
	Type      string `json:"type"`
	Transform string `json:"transform,omitempty"`
}

type SiteSelectors struct {
	Container string        `json:"container"`
	Item      string        `json:"item,omitempty"`
	Fields    []FieldConfig `json:"fields"`
}

type SiteConfig struct {
	Name          string        `json:"name"`
	URL           string        `json:"url"`
	Storage       string        `json:"storage"`
	Selectors     SiteSelectors `json:"selectors"`
	CheckInterval int           `json:"check_interval"`
}

type Config struct {
	Web WebConfig `json:"web"`

	Notification *struct {
		Service string                 `json:"service"`
		Config  map[string]interface{} `json:"config"`
	} `json:"notification,omitempty"` // 改为可选配置

	Sites []SiteConfig `json:"sites"`
}

type WebConfig struct {
	Port string `json:"port"` // 如 ":8080"
}

// LoadConfig 加载并校验配置
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg, json.DefaultOptionsV2()); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate 全局配置校验
func (c *Config) Validate() error {
	if len(c.Sites) == 0 {
		return errors.New("至少需要配置一个监控站点")
	}

	for i := range c.Sites {
		if err := c.Sites[i].validate(); err != nil {
			return fmt.Errorf("站点 #%d: %v", i+1, err)
		}
	}
	return nil
}

// validate 校验单个站点配置
func (s *SiteConfig) validate() error {
	if s.Name == "" {
		return errors.New("名称不能为空")
	}

	if _, err := url.ParseRequestURI(s.URL); err != nil {
		return fmt.Errorf("URL格式错误 (%s)", s.URL)
	}

	if s.CheckInterval < 0 {
		return errors.New("检查间隔不能为负数")
	}

	return s.Selectors.validate()
}

// validate 校验选择器配置
func (ss *SiteSelectors) validate() error {
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

// GetCheckInterval 获取检查间隔（需先调用Validate）
func (s *SiteConfig) GetCheckInterval() time.Duration {
	switch {
	case s.CheckInterval == 0:
		return 1 * time.Hour // 默认值
	case s.CheckInterval > 0:
		return time.Duration(s.CheckInterval) * time.Second
	default:
		panic("未校验的负数间隔") // Validate()应已拦截此情况
	}
}

/*
	// 1. 加载配置
	cfg, err := config.LoadConfig("config/default.json")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 2. 打印解析结果
	fmt.Println("✅ 配置加载成功")
	fmt.Printf("共加载 %d 个监控站点\n", len(cfg.Sites))

	for i, site := range cfg.Sites {
		fmt.Printf("\n🏷️ 站点 #%d\n", i+1)
		fmt.Printf("名称: %s\n", site.Name)
		fmt.Printf("URL: %s\n", site.URL)
		fmt.Printf("存储路径: %s\n", site.Storage)
		fmt.Printf("检查间隔: %v\n", site.GetCheckInterval())

		// 打印选择器配置
		fmt.Println("\n🔍 选择器配置:")
		fmt.Printf("容器: %q\n", site.Selectors.Container)
		fmt.Printf("列表项: %q\n", site.Selectors.Item)
		for _, field := range site.Selectors.Fields {
			fmt.Printf("  - 字段 %q: 选择器=%q", field.Name, field.Selector)
			if field.Attr != "" {
				fmt.Printf(", 属性=%q", field.Attr)
			}
			if field.Transform != "" {
				fmt.Printf(", 转换规则=%q", field.Transform)
			}
			fmt.Println()
		}
	}
*/
