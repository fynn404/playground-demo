# Go Playground

一个简单但功能强大的 Go 语言在线编程平台，支持代码编辑、执行和格式化。

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

## 系统要求

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

3. 运行服务器：
   ```bash
   go run main.go
   ```

4. 打开浏览器访问：`http://localhost:8081`

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

## 安全说明

当前版本主要用于开发和学习环境。在生产环境中使用时，建议添加以下安全措施：

- 代码执行超时限制
- 内存使用限制
- 网络访问限制
- Docker 容器化运行环境
- 用户输入验证和过滤
- 访问频率限制

## 技术栈

- 后端：Go
- 前端：HTML5, CSS3, JavaScript
- 代码编辑器：CodeMirror
- 代码格式化：gofmt

## 开源协议

MIT License