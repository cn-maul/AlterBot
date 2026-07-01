#!/bin/bash
set -e

echo "=== AlterBot 构建脚本 ==="

# 1. 构建前端
echo ""
echo "📦 [1/2] 构建前端..."
cd frontend
npm install --silent
npm run build
cd ..

# 2. 编译后端（嵌入前端产物）
echo ""
echo "🔨 [2/2] 编译后端..."
go build -o alterbot .

echo ""
echo "✅ 构建完成！"
echo "   运行: ./alterbot"
echo "   访问: http://localhost:8080"
