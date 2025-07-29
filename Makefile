cur_dir=$(shell pwd)
.PHONY: api

include $(cur_dir)/Makefile.env
run-api:
	go run $(cur_dir)/cmd/api/main.go


# 编译
build:
	@echo "Building..."
	@chmod +x scripts/build.sh
	@./scripts/build.sh

# 清理编译产物
clean:
	@echo "Cleaning..."
	@rm -rf output/

# 运行测试
test:
	@echo "Running tests..."
	@go test -v ./...

# 运行服务
run: build
	@echo "Starting service..."
	@chmod +x output/service.sh
	@cd output && ./service.sh start

# 停止服务
stop:
	@echo "Stopping service..."
	@cd output && ./service.sh stop

# 重启服务
restart: build
	@echo "Restarting service..."
	@cd output && ./service.sh restart

# 查看服务状态
status:
	@cd output && ./service.sh status


# 生成api文档
# 安装swag： go install github.com/swaggo/swag/cmd/swag@latest
build-tone-agent:
	swag init -g cmd/api/main.go