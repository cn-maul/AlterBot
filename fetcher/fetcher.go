package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type Fetcher struct {
	config *config // 持有私有配置的不可变副本
}

// New 创建Fetcher实例（线程安全）
func New(opts ...Option) *Fetcher {
	cfg := newDefaultConfig() // 深拷贝默认配置
	for _, opt := range opts {
		opt(cfg) // 应用用户配置
	}
	return &Fetcher{config: cfg}
}

// Fetch 执行HTTP请求
func (f *Fetcher) Fetch(url string) (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置User-Agent（唯一直接管理的Header）
	req.Header.Set("User-Agent", f.config.userAgent)

	// 执行请求（所有网络行为委托给http.Client）
	resp, err := f.config.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// 读取响应（限制10MB内存）
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return "", fmt.Errorf("读取失败: %w", err)
	}

	return string(body), nil
}

/*

	// 基础用法
	f1 := fetcher.New()
	result, err := f1.Fetch("https://example.com")

	// 自定义配置
	f2 := fetcher.New(
		fetcher.WithTimeout(10 * time.Second),
		fetcher.WithUserAgent("MyBot/2.0"),
	)
	result, _ = f2.Fetch("https://api.example.com")

	// 完全自定义Client
	customClient := &http.Client{Timeout: 3 * time.Second}
	f3 := fetcher.New(fetcher.WithClient(customClient))
	result, _ = f3.Fetch("https://internal.example.com")

*/
