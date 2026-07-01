.PHONY: all build run dev clean

all: build

# 编译后端 + 构建前端
build:
	cd frontend && npm install --silent && npm run build
	go build -o alterbot .
	@echo "✅ 构建完成！运行: ./alterbot"

# 本地运行（需先 build）
run:
	./alterbot

# 开发模式（启动后端，需要前端在另一个终端 npm run dev）
dev:
	go run .

# 前端开发服务器
dev-frontend:
	cd frontend && npm run dev

# 清理构建产物
clean:
	rm -f alterbot
	rm -rf frontend/dist
	rm -f alterbot.db
