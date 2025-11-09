# Go Chat Backend Makefile
# 简化开发和部署流程

.PHONY: help build run test clean docker-build docker-run dev setup

# 默认目标
help:
	@echo "Go Chat Backend - 可用命令:"
	@echo ""
	@echo "  setup        - 初始化项目环境"
	@echo "  dev          - 开发模式运行"
	@echo "  build        - 编译应用"
	@echo "  run          - 运行应用"
	@echo "  test         - 运行测试"
	@echo "  clean        - 清理编译文件"
	@echo "  docker-build - 构建Docker镜像"
	@echo "  docker-run   - 运行Docker容器"
	@echo "  deploy       - 使用Docker Compose部署"
	@echo "  logs         - 查看应用日志"
	@echo "  stop         - 停止所有服务"
	@echo ""

# 初始化项目环境
setup:
	@echo "📦 初始化项目环境..."
	@if [ ! -f .env ]; then cp .env.example .env; echo "✅ 已创建 .env 文件"; fi
	@mkdir -p logs
	@go mod tidy
	@echo "✅ 环境初始化完成"
	@echo "⚠️  请编辑 .env 文件配置您的参数"

# 开发模式运行
dev: setup
	@echo "🚀 开发模式启动..."
	@go run main.go

# 编译应用
build:
	@echo "🔨 编译应用..."
	@go build -o bin/go-chat-backend .
	@echo "✅ 编译完成: bin/go-chat-backend"

# 运行应用
run: build
	@echo "🚀 运行应用..."
	@./bin/go-chat-backend

# 运行测试
test:
	@echo "🧪 运行测试..."
	@go test -v ./...

# 清理编译文件
clean:
	@echo "🧹 清理编译文件..."
	@rm -rf bin/
	@rm -rf logs/*.log
	@go clean
	@echo "✅ 清理完成"

# 构建Docker镜像
docker-build:
	@echo "🐳 构建Docker镜像..."
	@docker build -t go-chat-backend .
	@echo "✅ Docker镜像构建完成"

# 运行Docker容器
docker-run: docker-build
	@echo "🐳 运行Docker容器..."
	@docker run --rm -p 8080:8080 --env-file .env go-chat-backend

# 使用Docker Compose部署
deploy:
	@echo "🚀 使用Docker Compose部署..."
	@chmod +x deploy.sh
	@./deploy.sh

# 查看应用日志
logs:
	@echo "📋 查看应用日志..."
	@if [ -f docker-compose.yml ]; then \
		docker-compose logs -f app; \
	else \
		tail -f logs/app.log; \
	fi

# 停止所有服务
stop:
	@echo "⏹️  停止所有服务..."
	@if [ -f docker-compose.yml ]; then \
		docker-compose down; \
		echo "✅ Docker Compose服务已停止"; \
	else \
		echo "ℹ️  没有找到docker-compose.yml文件"; \
	fi

# 重启服务
restart: stop deploy

# 查看服务状态
status:
	@echo "📊 服务状态:"
	@if [ -f docker-compose.yml ]; then \
		docker-compose ps; \
	else \
		echo "ℹ️  没有找到docker-compose.yml文件"; \
	fi

# 数据库迁移（如果需要手动运行）
migrate:
	@echo "🗃️  运行数据库迁移..."
	@go run scripts/migrate.go

# 生成API文档
docs:
	@echo "📚 生成API文档..."
	@echo "ℹ️  请查看 README.md 中的API文档部分"

# 代码格式化
fmt:
	@echo "🎨 格式化代码..."
	@go fmt ./...
	@echo "✅ 代码格式化完成"

# 代码检查
lint:
	@echo "🔍 运行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint 未安装，跳过代码检查"; \
		echo "   安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 安全检查
sec:
	@echo "🔒 运行安全检查..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec 未安装，跳过安全检查"; \
		echo "   安装命令: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# 性能基准测试
bench:
	@echo "⚡ 运行性能基准测试..."
	@go test -bench=. ./...

# 覆盖率测试
coverage:
	@echo "📊 运行覆盖率测试..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆盖率报告已生成: coverage.html"

# 依赖更新
update:
	@echo "📦 更新依赖..."
	@go get -u ./...
	@go mod tidy
	@echo "✅ 依赖更新完成"

# 完整检查（格式化 + 检查 + 测试）
check: fmt lint test
	@echo "✅ 所有检查通过"

# 生产构建
release: clean check build
	@echo "🎉 生产版本构建完成"
