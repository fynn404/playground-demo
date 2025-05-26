#!/bin/sh

# 接收代码作为参数
CODE="$1"

# 创建临时目录
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR" || exit 1

# 写入代码到文件
echo "$CODE" > main.go

# 编译并运行代码，设置超时
timeout 5s go run main.go

# 获取退出状态
EXIT_CODE=$?

# 清理临时文件
cd / && rm -rf "$TEMP_DIR"

# 根据退出状态返回适当的消息
if [ $EXIT_CODE -eq 124 ]; then
    echo "Error: Execution timeout after 5 seconds"
    exit 124
elif [ $EXIT_CODE -ne 0 ]; then
    exit $EXIT_CODE
fi 