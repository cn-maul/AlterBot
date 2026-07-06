package notify

import (
	"fmt"

	serverchan_sdk "github.com/easychen/serverchan-sdk-golang"
)

func init() {
	Register("serverchan", newServerChanNotifier)
}

func newServerChanNotifier(config map[string]interface{}) (Notifier, error) {
	sendkey, _ := config["sendkey"].(string)
	if sendkey == "" {
		return nil, fmt.Errorf("缺少必需的 sendkey 参数")
	}

	return &serverChanNotifier{
		sendkey: sendkey,
	}, nil
}

type serverChanNotifier struct {
	sendkey string
}

func (s *serverChanNotifier) Send(title, content string) error {
	resp, err := serverchan_sdk.ScSend(s.sendkey, title, content, nil)
	if err != nil {
		return fmt.Errorf("Server酱推送失败: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("Server酱返回错误: %s (code=%d)", resp.Message, resp.Code)
	}
	return nil
}

func (s *serverChanNotifier) Name() string {
	return "serverchan"
}
