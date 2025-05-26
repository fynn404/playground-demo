# Go Playground

一个安全可靠的 Go 语言在线编程平台，支持代码编辑、执行和格式化。采用 Docker 容器化技术，确保代码执行的安全性。

## 功能特点

- 🎨 现代化的用户界面
  - 支持代码语法高亮
  - 自动括号匹配
  - 行号显示
  - 自动缩进

- ⚡ 实时代码执行
  - 在线运行 Go 代码
  - 实时显示执行结果
  - 清晰的错误提示

- 🔧 代码格式化
  - 使用 `gofmt` 进行标准格式化
  - 保持代码风格一致性
  - 自动对齐和缩进

- ⌨️ 快捷键支持
  - 运行代码：`Ctrl/Cmd + Enter`
  - 格式化代码：`Ctrl/Cmd + S`
  - 智能缩进：`Tab`

- 🛡️ 安全特性
  - Docker 容器化执行
  - 资源使用限制
  - 网络访问限制
  - 系统调用限制
  - 代码执行超时控制
  - IP 速率限制

## 系统要求

- Docker 20.10 或更高版本
- Go 1.16 或更高版本
- 现代浏览器（Chrome、Firefox、Safari、Edge 等）

## 安装和运行

1. 克隆仓库：
   ```bash
   git clone <repository-url>
   ```

2. 进入项目目录：
   ```bash
   cd playground-demo
   ```

3. 使用 Docker Compose 启动服务：
   ```bash
   docker-compose up -d
   ```

4. 打开浏览器访问：`http://localhost:8081`

## 安全措施

### 1. 容器化执行
- 每个代码片段在独立的 Docker 容器中执行
- 容器在执行完成后自动销毁
- 使用非特权用户运行代码

### 2. 资源限制
- 内存限制：100MB
- CPU 限制：0.5 核
- 最大进程数：20
- 文件描述符限制：64
- 代码大小限制：10KB
- 输出大小限制：1MB

### 3. 网络限制
- 完全禁用网络访问
- 无法访问外部资源
- 无法建立网络连接

### 4. 执行限制
- 代码执行超时：5 秒
- 禁止危险系统调用
- 禁止文件系统操作
- 禁止执行外部命令

### 5. API 访问控制
- IP 基于令牌桶的速率限制
- 初始令牌数：3
- 令牌恢复率：每秒 0.5 个
- 最大令牌数：3

### 6. 代码安全检查
禁止导入以下包：
- `os/exec`: 禁止执行系统命令
- `syscall`: 禁止系统调用
- `unsafe`: 禁止不安全操作
- `net`: 禁止网络操作
- `os`: 禁止文件系统操作
- `plugin`: 禁止加载插件
- `debug`: 禁止调试功能

## 使用指南

### 基本使用

1. 在编辑器中输入 Go 代码
2. 点击 "Run Code" 按钮或使用快捷键 `Ctrl/Cmd + Enter` 运行代码
3. 在下方查看执行结果或错误信息

### 代码格式化

1. 输入需要格式化的代码
2. 点击 "Format Code" 按钮或使用快捷键 `Ctrl/Cmd + S`
3. 代码将自动按照 Go 标准格式进行格式化

### 示例代码

```go
package main

import "fmt"

func main() {
    // 打印 Hello World
    fmt.Println("Hello, Go Playground!")
    
    // 简单的循环示例
    for i := 1; i <= 5; i++ {
        fmt.Printf("%d 的平方是 %d\n", i, i*i)
    }
}
```

### 快捷键列表

- `Ctrl/Cmd + Enter`: 运行代码
- `Ctrl/Cmd + S`: 格式化代码
- `Tab`: 插入制表符或缩进选中的代码块
- `Shift + Tab`: 减少缩进

## API 接口

### 1. 运行代码

- 端点：`POST /run`
- 请求体：
  ```json
  {
    "code": "你的 Go 代码"
  }
  ```
- 响应：
  ```json
  {
    "output": "程序输出",
    "error": "错误信息（如果有）"
  }
  ```
- 速率限制：每个 IP 每秒最多 3 个请求

### 2. 格式化代码

- 端点：`POST /format`
- 请求体：
  ```json
  {
    "code": "需要格式化的代码"
  }
  ```
- 响应：
  ```json
  {
    "formattedCode": "格式化后的代码",
    "error": "错误信息（如果有）"
  }
  ```
- 速率限制：每个 IP 每秒最多 3 个请求

## 技术栈

- 后端：Go
- 前端：HTML5, CSS3, JavaScript
- 代码编辑器：CodeMirror
- 代码格式化：gofmt
- 容器化：Docker
- 安全：seccomp, Docker 容器限制
- 速率限制：令牌桶算法

## 开源协议

MIT License