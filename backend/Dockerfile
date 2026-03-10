# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 设置Go代理
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# 安装依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download || go mod tidy

# 复制源代码
COPY . .

# 整理依赖并构建
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /campus-server ./cmd/server

# 运行阶段
FROM alpine:3.18

WORKDIR /app

# 安装ca证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 从构建阶段复制二进制文件
COPY --from=builder /campus-server /app/campus-server
COPY --from=builder /app/config.yaml /app/config.yaml

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080

# 运行
CMD ["/app/campus-server", "-config", "/app/config.yaml"]