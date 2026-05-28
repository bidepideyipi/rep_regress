.PHONY: help build run clean deps test fmt lint

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
APP_NAME := slot-game
BUILD_DIR := build
CONFIG_FILE := config.yaml

## help: 显示帮助信息
help:
	@echo "可用命令:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

## build: 编译项目
build:
	@echo "编译项目..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) main.go

## run: 运行项目
run:
	@echo "运行项目..."
	@go run main.go -config=$(CONFIG_FILE)

## run-debug: 调试模式运行
run-debug:
	@echo "调试模式运行项目..."
	@go run main.go -config=$(CONFIG_FILE)

## deps: 下载依赖
deps:
	@echo "下载依赖..."
	@go mod download
	@go mod tidy

## clean: 清理编译文件
clean:
	@echo "清理编译文件..."
	@rm -rf $(BUILD_DIR)
	@go clean

## test: 运行测试
test:
	@echo "运行测试..."
	@go test -v ./...

## fmt: 格式化代码
fmt:
	@echo "格式化代码..."
	@go fmt ./...

## lint: 代码检查
lint:
	@echo "代码检查..."
	@golangci-lint run

## dev: 开发环境启动
dev: deps fmt
	@echo "启动开发环境..."
	@go run main.go -config=$(CONFIG_FILE)

## docker-build: 构建Docker镜像
docker-build:
	@echo "构建Docker镜像..."
	@docker build -t $(APP_NAME):latest .

## docker-run: 运行Docker容器
docker-run:
	@echo "运行Docker容器..."
	@docker run -p 8080:8080 --name $(APP_NAME) $(APP_NAME):latest

## install: 安装项目到本地
install: build
	@echo "安装项目到本地..."
	@cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/