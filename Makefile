.PHONY: all build run dev clean release

all: build

# 编译后端 + 构建前端（纯 Go，无 CGo）
build:
	cd frontend && npm install --silent && npm run build
	CGO_ENABLED=0 go build -ldflags="-s -w" -o alterbot .
	@echo "✅ 构建完成！运行: ./alterbot"

# 本地运行（需先 build）
run:
	./alterbot

# 开发模式（启动后端，需要前端在另一个终端 npm run dev）
dev:
	CGO_ENABLED=0 go run .

# 前端开发服务器
dev-frontend:
	cd frontend && npm run dev

# 跨平台发布
release:
	cd frontend && npm install --silent && npm run build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o alterbot-linux-amd64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o alterbot-windows-amd64.exe .
	@echo "✅ 发布包构建完成"

# 清理构建产物
clean:
	rm -f alterbot alterbot-linux-amd64 alterbot-windows-amd64.exe
	rm -rf frontend/dist
	rm -f alterbot.db