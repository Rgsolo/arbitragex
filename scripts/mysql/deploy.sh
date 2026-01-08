#!/bin/bash

# ================================================================================
# ArbitrageX MySQL 快速部署脚本
# ================================================================================
# 版本: v1.0.0
# 创建日期: 2026-01-08
# 维护人: yangyangyang
# 描述: 快速部署 MySQL 数据库（使用 Docker）
# ================================================================================

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ================================================================================
# 配置变量
# ================================================================================

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MYSQL_CONTAINER_NAME="arbitragex-mysql"
MYSQL_ROOT_PASSWORD="root_password"
MYSQL_DATABASE="arbitragex"
MYSQL_USER="arbitragex_user"
MYSQL_PASSWORD="ArbitrageX2025!"
MYSQL_PORT=3306

# ================================================================================
# 辅助函数
# ================================================================================

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo ""
    echo "================================================================================"
    echo "$1"
    echo "================================================================================"
    echo ""
}

# ================================================================================
# 检查 Docker
# ================================================================================

check_docker() {
    print_header "检查 Docker 环境"

    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi

    print_success "Docker 已安装: $(docker --version)"

    if ! command -v docker-compose &> /dev/null; then
        print_warning "docker-compose 未安装，将使用 docker 命令"
        USE_DOCKER_COMPOSE=0
    else
        print_success "docker-compose 已安装: $(docker-compose --version)"
        USE_DOCKER_COMPOSE=1
    fi
}

# ================================================================================
# 检查端口
# ================================================================================

check_port() {
    print_header "检查端口 ${MYSQL_PORT}"

    if lsof -Pi :${MYSQL_PORT} -sTCP:LISTEN -t >/dev/null 2>&1; then
        print_warning "端口 ${MYSQL_PORT} 已被占用"
        read -p "是否终止占用该端口的进程? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            lsof -ti:${MYSQL_PORT} | xargs kill -9
            print_success "已终止占用端口的进程"
        else
            print_error "端口 ${MYSQL_PORT} 被占用，无法启动 MySQL"
            exit 1
        fi
    else
        print_success "端口 ${MYSQL_PORT} 可用"
    fi
}

# ================================================================================
# 创建数据目录
# ================================================================================

create_data_dir() {
    print_header "创建数据目录"

    mkdir -p "${PROJECT_ROOT}/data/mysql"
    print_success "数据目录已创建: ${PROJECT_ROOT}/data/mysql"
}

# ================================================================================
# 启动 MySQL（使用 docker-compose）
# ================================================================================

start_mysql_compose() {
    print_header "启动 MySQL 容器（使用 docker-compose）"

    cd "${PROJECT_ROOT}"

    # 检查 docker-compose.yml 是否存在
    if [ ! -f "docker-compose.yml" ]; then
        print_warning "docker-compose.yml 不存在，将创建临时配置"

        cat > docker-compose.yml <<EOF
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: ${MYSQL_CONTAINER_NAME}
    restart: always
    ports:
      - "${MYSQL_PORT}:3306"
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: ${MYSQL_DATABASE}
      MYSQL_USER: ${MYSQL_USER}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
      TZ: Asia/Shanghai
    volumes:
      - ./data/mysql:/var/lib/mysql
      - ./scripts/mysql:/docker-entrypoint-initdb.d
      - ./config/mysql.cnf:/etc/mysql/conf.d/custom.cnf
    command:
      - --character-set-server=utf8mb4
      - --collation-server=utf8mb4_unicode_ci
      - --default-authentication-plugin=mysql_native_password
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 3
EOF
    fi

    # 启动容器
    docker-compose up -d mysql

    # 等待容器启动
    print_info "等待 MySQL 容器启动..."
    sleep 10

    # 检查容器状态
    if docker ps | grep -q ${MYSQL_CONTAINER_NAME}; then
        print_success "MySQL 容器启动成功"
    else
        print_error "MySQL 容器启动失败"
        exit 1
    fi
}

# ================================================================================
# 启动 MySQL（使用 docker）
# ================================================================================

start_mysql_docker() {
    print_header "启动 MySQL 容器（使用 docker）"

    # 检查容器是否已存在
    if docker ps -a | grep -q ${MYSQL_CONTAINER_NAME}; then
        print_warning "容器 ${MYSQL_CONTAINER_NAME} 已存在"
        read -p "是否删除并重新创建? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker stop ${MYSQL_CONTAINER_NAME} 2>/dev/null || true
            docker rm ${MYSQL_CONTAINER_NAME} 2>/dev/null || true
            print_success "已删除旧容器"
        else
            docker start ${MYSQL_CONTAINER_NAME}
            print_success "已启动现有容器"
            return
        fi
    fi

    # 启动容器
    docker run --name ${MYSQL_CONTAINER_NAME} \
        -e MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD} \
        -e MYSQL_DATABASE=${MYSQL_DATABASE} \
        -e MYSQL_USER=${MYSQL_USER} \
        -e MYSQL_PASSWORD=${MYSQL_PASSWORD} \
        -e TZ=Asia/Shanghai \
        -p ${MYSQL_PORT}:3306 \
        -v "${PROJECT_ROOT}/data/mysql:/var/lib/mysql" \
        -v "${PROJECT_ROOT}/scripts/mysql:/docker-entrypoint-initdb.d" \
        -v "${PROJECT_ROOT}/config/mysql.cnf:/etc/mysql/conf.d/custom.cnf" \
        -d mysql:8.0 \
        --character-set-server=utf8mb4 \
        --collation-server=utf8mb4_unicode_ci \
        --default-authentication-plugin=mysql_native_password

    # 等待容器启动
    print_info "等待 MySQL 容器启动..."
    sleep 10

    # 检查容器状态
    if docker ps | grep -q ${MYSQL_CONTAINER_NAME}; then
        print_success "MySQL 容器启动成功"
    else
        print_error "MySQL 容器启动失败"
        exit 1
    fi
}

# ================================================================================
# 验证数据库
# ================================================================================

verify_database() {
    print_header "验证数据库"

    # 等待 MySQL 完全启动
    print_info "等待 MySQL 完全启动..."
    for i in {1..30}; do
        if docker exec ${MYSQL_CONTAINER_NAME} mysqladmin ping -h localhost --silent; then
            print_success "MySQL 已完全启动"
            break
        fi
        if [ $i -eq 30 ]; then
            print_error "MySQL 启动超时"
            exit 1
        fi
        sleep 2
    done

    # 连接数据库
    print_info "验证数据库连接..."
    if docker exec ${MYSQL_CONTAINER_NAME} mysql -u${MYSQL_USER} -p${MYSQL_PASSWORD} ${MYSQL_DATABASE} -e "SELECT 'Database connection successful!' AS Message;" 2>/dev/null; then
        print_success "数据库连接成功"
    else
        print_error "数据库连接失败"
        exit 1
    fi

    # 查看表
    print_info "查看数据库表..."
    TABLE_COUNT=$(docker exec ${MYSQL_CONTAINER_NAME} mysql -u${MYSQL_USER} -p${MYSQL_PASSWORD} ${MYSQL_DATABASE} -e "SHOW TABLES;" 2>/dev/null | wc -l)
    print_success "数据库包含 $((TABLE_COUNT - 1)) 个表"

    # 查看系统配置
    print_info "查看系统配置..."
    docker exec ${MYSQL_CONTAINER_NAME} mysql -u${MYSQL_USER} -p${MYSQL_PASSWORD} ${MYSQL_DATABASE} -e "SELECT config_key, config_value, description FROM system_config;" 2>/dev/null
}

# ================================================================================
# 显示连接信息
# ================================================================================

show_connection_info() {
    print_header "数据库连接信息"

    echo "数据库名称: ${MYSQL_DATABASE}"
    echo "用户名: ${MYSQL_USER}"
    echo "密码: ${MYSQL_PASSWORD}"
    echo "端口: ${MYSQL_PORT}"
    echo "容器名: ${MYSQL_CONTAINER_NAME}"
    echo ""
    echo "连接字符串:"
    echo "  命令行: mysql -h localhost -P ${MYSQL_PORT} -u ${MYSQL_USER} -p${MYSQL_PASSWORD} ${MYSQL_DATABASE}"
    echo "  Docker:  docker exec -it ${MYSQL_CONTAINER_NAME} mysql -u${MYSQL_USER} -p${MYSQL_PASSWORD} ${MYSQL_DATABASE}"
    echo ""
    echo "Go (go-zero):"
    echo "  Mysql:"
    echo "    DataSource: \"${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(localhost:${MYSQL_PORT})/${MYSQL_DATABASE}?charset=utf8mb4&parseTime=true\""
    echo ""
}

# ================================================================================
# 主函数
# ================================================================================

main() {
    print_header "ArbitrageX MySQL 数据库快速部署"

    # 检查 Docker
    check_docker

    # 检查端口
    check_port

    # 创建数据目录
    create_data_dir

    # 启动 MySQL
    if [ $USE_DOCKER_COMPOSE -eq 1 ]; then
        start_mysql_compose
    else
        start_mysql_docker
    fi

    # 验证数据库
    verify_database

    # 显示连接信息
    show_connection_info

    print_success "MySQL 数据库部署完成！"
    echo ""
    print_info "常用命令:"
    echo "  查看日志: docker logs -f ${MYSQL_CONTAINER_NAME}"
    echo "  停止容器: docker stop ${MYSQL_CONTAINER_NAME}"
    echo "  启动容器: docker start ${MYSQL_CONTAINER_NAME}"
    echo "  删除容器: docker rm -f ${MYSQL_CONTAINER_NAME}"
    echo ""
}

# 运行主函数
main
