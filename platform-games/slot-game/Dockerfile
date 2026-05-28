FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git make

# 复制go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN go build -o slot-game main.go

# 运行阶段
FROM alpine:latest

# 安装必要的库
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建工作目录
WORKDIR /app

# 从编译阶段复制二进制文件
COPY --from=builder /app/slot-game .
COPY --from=builder /app/config.yaml .

# 创建日志目录
RUN mkdir -p logs

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./slot-game", "-config=config.yaml"]