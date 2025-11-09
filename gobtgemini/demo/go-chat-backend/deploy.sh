#!/bin/bash

# Go Chat Backend 部署脚本
# 作者: Go Chat Team
# 版本: 1.0.0

set -e  # 遇到错误就退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 创建环境文件
create_env_file() {
    if [ ! -f .env ]; then
        log_info "创建 .env 文件..."
        cp .env.example .env
        log_warning "请编辑 .env 文件并配置您的实际参数"
        echo "特别是以下参数："
        echo "- LLM_API_KEY: 您的大模型API密钥"
        echo "- JWT_SECRET: 一个安全的随机字符串"
        echo "- DB_PASSWORD: 数据库密码"
    else
        log_info ".env 文件已存在，跳过创建"
    fi
}

# 创建目录
create_directories() {
    log_info "创建必要的目录..."
    mkdir -p logs
    mkdir -p ssl  # 用于SSL证书（如果需要）
    log_success "目录创建完成"
}

# 构建和启动服务
start_services() {
    log_info "构建和启动服务..."
    
    # 停止现有服务（如果有）
    docker-compose down 2>/dev/null || true
    
    # 构建和启动服务
    docker-compose up --build -d
    
    log_success "服务启动成功"
}

# 等待服务就绪
wait_for_services() {
    log_info "等待服务就绪..."
    
    # 等待PostgreSQL
    log_info "等待 PostgreSQL..."
    timeout=60
    while ! docker-compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; do
        if [ $timeout -le 0 ]; then
            log_error "PostgreSQL 启动超时"
            exit 1
        fi
        timeout=$((timeout-1))
        sleep 1
    done
    
    # 等待Chroma
    log_info "等待 Chroma..."
    timeout=60
    while ! curl -f http://localhost:8000/api/v1/heartbeat >/dev/null 2>&1; do
        if [ $timeout -le 0 ]; then
            log_error "Chroma 启动超时"
            exit 1
        fi
        timeout=$((timeout-1))
        sleep 1
    done
    
    # 等待应用
    log_info "等待应用..."
    timeout=60
    while ! curl -f http://localhost:8080/health >/dev/null 2>&1; do
        if [ $timeout -le 0 ]; then
            log_error "应用启动超时"
            exit 1
        fi
        timeout=$((timeout-1))
        sleep 1
    done
    
    log_success "所有服务已就绪"
}

# 运行测试
run_tests() {
    log_info "运行基本测试..."
    
    # 测试健康检查
    if curl -f http://localhost:8080/health >/dev/null 2>&1; then
        log_success "健康检查通过"
    else
        log_error "健康检查失败"
        return 1
    fi
    
    # 测试注册 API
    response=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/api/v1/auth/register \
        -H "Content-Type: application/json" \
        -d '{"username":"testuser","email":"test@example.com","password":"password123"}' 2>/dev/null)
    
    if [ "$response" = "201" ] || [ "$response" = "409" ]; then
        log_success "注册API测试通过"
    else
        log_warning "注册API测试未通过，响应码: $response"
    fi
}

# 显示服务状态
show_status() {
    log_info "服务状态:"
    docker-compose ps
    
    echo ""
    log_info "服务地址:"
    echo "- 应用: http://localhost:8080"
    echo "- API文档: http://localhost:8080/health"
    echo "- PostgreSQL: localhost:5432"
    echo "- Chroma: http://localhost:8000"
    echo "- Nginx (可选): http://localhost"
    
    echo ""
    log_info "常用命令:"
    echo "- 查看日志: docker-compose logs -f app"
    echo "- 停止服务: docker-compose down"
    echo "- 重启服务: docker-compose restart"
    echo "- 更新服务: docker-compose up --build -d"
}

# 主函数
main() {
    echo "================================================"
    echo "  Go Chat Backend 部署脚本"
    echo "================================================"
    echo ""
    
    check_dependencies
    create_env_file
    create_directories
    start_services
    wait_for_services
    run_tests
    
    echo ""
    echo "================================================"
    log_success "\u90e8\u7f72\u5b8c\u6210\uff01"
    echo "================================================"
    echo ""
    
    show_status
}

# 运行主函数
main "$@"
