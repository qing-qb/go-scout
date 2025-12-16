# 第一阶段：构建阶段 (Builder)
# 使用官方 Go 镜像作为构建环境
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制依赖文件 go.mod 和 go.sum (如果存在)
COPY go.mod ./
# RUN go mod download

# 复制源代码
COPY . .

# 静态编译 Go 程序 (禁用 CGO，确保在轻量级容器中运行)
RUN CGO_ENABLED=0 GOOS=linux go build -o gscout ./cmd/gscout/main.go

# ---

# 第二阶段：运行阶段 (Runner)
# 使用最精简的 Alpine 镜像
FROM alpine:latest

# 安装基础证书 (如果扫描 HTTPS 需要)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建阶段只复制编译好的二进制文件
COPY --from=builder /app/gscout .

# 赋予执行权限
RUN chmod +x ./gscout

# 设置入口点
ENTRYPOINT ["./gscout"]

# 默认参数 (可以通过 docker run 覆盖)
CMD ["-h"]
