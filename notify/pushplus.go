package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func init() {
	Register("pushplus", newPushPlusNotifier)
}

func newPushPlusNotifier(config map[string]interface{}) (Notifier, error) {
	token, _ := config["token"].(string)
	if token == "" {
		return nil, fmt.Errorf("缺少必需的token参数")
	}

	channel, _ := config["channel"].(string)
	return &pushPlusNotifier{
		token:   token,
		channel: channel,
	}, nil
}

type pushPlusNotifier struct {
	token   string
	channel string
}

func (p *pushPlusNotifier) Send(title, content string) error {
	payload := map[string]interface{}{
		"token":   p.token,
		"title":   title,
		"content": content + "\n",
	}
	if p.channel != "" {
		payload["channel"] = p.channel
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("JSON编码失败: %w", err)
	}

	resp, err := http.Post(
		"http://www.pushplus.plus/send",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}

func (p *pushPlusNotifier) Name() string {
	return "pushplus"
}
