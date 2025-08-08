# AlterBot 网页更新通知

## 1、使用
在你的项目根目录创建目录`config`用于存放监视器的配置文件，创建配置文件`default.json`，编译时需要设置`GOEXPERIMENT=jsonv2,greenteagc`
```json
{
  "notification": {
    "service": "pushplus",
    "config": {
      "token": "yourtoken",
      "channel": "mail"
    }
  },
  "sites": [
    {
      "name": "招录公告",
      "url": "https://xxx.cn//zlgg/",
      "storage": "storage/xxx.json",
      "selectors": {
        "container": "div.hap_infoBox",
        "item": "div.hap_infoOne",
        "fields": [
          {
            "name": "title",
            "selector": "a",
            "type": "text"
          },
          {
            "name": "url",
            "selector": "a",
            "attr": "href",
            "type": "attr"
          },
          {
            "name": "date",
            "selector": "span.hap_infoDate",
            "type": "text",
            "transform": "trim([], '[]')"
          }
        ]
      },
      "check_interval": 30
    }
  ]
}
```

首先加载配置文件
```go
cfg, err := config.LoadConfig("config/default.json")
if err != nil {
log.Fatalf("加载配置失败: %v", err)
}
```
初始化推送服务
```go
if cfg.Notification != nil {
		if err := notify.InitGlobalNotifier(
			cfg.Notification.Service,
			cfg.Notification.Config,
		); err != nil {
			log.Fatal("推送服务初始化失败:", err)
		}
	}
```
启动监控goroutine
```go
var wg sync.WaitGroup
	stopCh := make(chan struct{})
	
	for _, site := range cfg.Sites {
		wg.Go(func() { // Go 1.25新语法
			monitor.Start(&site, stopCh)
		})
	}
```
等待中断信号,关闭所有监控器
```go
sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	
	close(stopCh)
	wg.Wait()
	log.Println("所有监控器已停止")
```
