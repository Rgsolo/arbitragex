# ArbitrageX Makefile
# 用于构建、运行、测试项目

.PHONY: all build run test clean fmt lint deps verify-stage verify-quick verify-full check-startup

# 默认目标
all: deps build

# 安装依赖
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# 格式化代码
fmt:
	@echo "Formatting code..."
	gofmt -w .
	goimports -w .

# 代码检查
lint:
	@echo "Running linters..."
	golangci-lint run

# 构建所有服务
build: fmt
	@echo "Building services..."
	mkdir -p bin
	go build -o bin/price-monitor ./restful/price
	go build -o bin/arbitrage-engine ./restful/engine
	go build -o bin/trade-executor ./restful/trade
	@echo "Build complete!"

# 构建单个服务
build-price:
	@echo "Building price monitor service..."
	go build -o bin/price-monitor ./restful/price

build-engine:
	@echo "Building arbitrage engine service..."
	go build -o bin/arbitrage-engine ./restful/engine

build-trade:
	@echo "Building trade executor service..."
	go build -o bin/trade-executor ./restful/trade

# 运行服务
run-price: build-price
	@echo "Running price monitor service..."
	./bin/price-monitor -f restful/price/etc/price-api.yaml

run-engine: build-engine
	@echo "Running arbitrage engine service..."
	./bin/arbitrage-engine -f restful/engine/etc/engine-api.yaml

run-trade: build-trade
	@echo "Running trade executor service..."
	./bin/trade-executor -f restful/trade/etc/trade-api.yaml

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./...

# 运行测试并显示覆盖率
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 运行基准测试
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# 清理构建文件
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "Clean complete!"

# 生成 go-zero 代码
generate-api:
	@echo "Generating API code..."
	goctl api go -api api/price.api -dir ./restful/price -style go_zero
	goctl api go -api api/engine.api -dir ./restful/engine -style go_zero
	goctl api go -api api/trade.api -dir ./restful/trade -style go_zero
	@echo "API code generation complete!"

# 验证当前阶段（完整验证）
verify-stage:
	@echo "Running stage verification..."
	@bash scripts/verify-stage.sh

# 快速验证（编译+测试）
verify-quick: build test
	@echo "Quick verification passed!"

# 完整验证（包括 Docker）
verify-full: verify-stage docker-build
	@echo "Full verification passed!"

# 检查服务启动
check-startup:
	@echo "Checking service startup..."
	@docker-compose up -d
	@sleep 5
	@docker-compose ps
	@echo "Startup check complete!"

# Docker 相关命令
docker-build:
	@echo "Building Docker images..."
	docker-compose build

docker-up:
	@echo "Starting Docker containers..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose down

docker-logs:
	@echo "Showing Docker logs..."
	docker-compose logs -f

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  all           - Install dependencies and build all services"
	@echo "  deps          - Install dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linters"
	@echo "  build         - Build all services"
	@echo "  build-price   - Build price monitor service"
	@echo "  build-engine  - Build arbitrage engine service"
	@echo "  build-trade   - Build trade executor service"
	@echo "  run-price     - Run price monitor service"
	@echo "  run-engine    - Run arbitrage engine service"
	@echo "  run-trade     - Run trade executor service"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  bench         - Run benchmarks"
	@echo "  clean         - Clean build files"
	@echo "  generate-api  - Generate go-zero API code"
	@echo "  verify-stage  - Run complete stage verification"
	@echo "  verify-quick  - Run quick verification (build + test)"
	@echo "  verify-full   - Run full verification (including Docker)"
	@echo "  check-startup - Check service startup"
	@echo "  docker-build  - Build Docker images"
	@echo "  docker-up     - Start Docker containers"
	@echo "  docker-down   - Stop Docker containers"
	@echo "  docker-logs   - Show Docker logs"
	@echo "  help          - Show this help message"
