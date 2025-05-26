FROM golang:1.21-alpine

# 安装必要的工具
RUN apk add --no-cache gcc musl-dev

# 创建非特权用户
RUN adduser -D -u 10001 coderunner

# 创建工作目录
WORKDIR /sandbox

# 复制运行脚本
COPY runner.sh /sandbox/
RUN chmod +x /sandbox/runner.sh

# 切换到非特权用户
USER coderunner

# 设置环境变量
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

# 入口点设置为运行脚本
ENTRYPOINT ["/sandbox/runner.sh"] 